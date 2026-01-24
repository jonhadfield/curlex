package executor

import (
	"context"
	"time"

	"curlex/internal/models"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	BackoffType   string // "exponential" or "linear"
	RetryOnStatus []int  // Status codes to retry on
}

// shouldRetry determines if a request should be retried based on the status code
func shouldRetry(statusCode int, retryOnStatus []int) bool {
	// If no specific status codes configured, don't retry
	if len(retryOnStatus) == 0 {
		return false
	}

	// Check if status code is in the retry list
	for _, code := range retryOnStatus {
		if statusCode == code {
			return true
		}
	}

	return false
}

// calculateDelay calculates the delay before the next retry attempt
func calculateDelay(attempt int, initialDelay time.Duration, backoffType string) time.Duration {
	if initialDelay == 0 {
		initialDelay = 1 * time.Second // Default to 1 second
	}

	switch backoffType {
	case "exponential":
		// Exponential backoff: delay * 2^attempt
		multiplier := 1 << uint(attempt) // 2^attempt
		return initialDelay * time.Duration(multiplier)
	case "linear":
		// Linear backoff: delay * attempt
		return initialDelay * time.Duration(attempt+1)
	default:
		// Default to exponential
		multiplier := 1 << uint(attempt)
		return initialDelay * time.Duration(multiplier)
	}
}

// ExecuteWithRetry executes a test with retry logic
func (e *Executor) ExecuteWithRetry(ctx context.Context, test models.Test) (*models.TestResult, error) {
	var lastResult *models.TestResult
	var lastErr error

	maxAttempts := test.Retries + 1 // Original attempt + retries
	if maxAttempts <= 1 {
		// No retries configured, execute normally
		return e.Execute(ctx, test)
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Execute the test
		result, err := e.Execute(ctx, test)

		// If execution succeeded, return immediately
		if err == nil && result.Success {
			return result, nil
		}

		// Save last result and error
		lastResult = result
		lastErr = err

		// Check if we should retry
		isLastAttempt := attempt == maxAttempts-1
		if isLastAttempt {
			break // Don't sleep after last attempt
		}

		// Determine if we should retry based on status code
		shouldRetryRequest := false
		if result != nil && len(test.RetryOnStatus) > 0 {
			shouldRetryRequest = shouldRetry(result.StatusCode, test.RetryOnStatus)
		} else if result != nil && !result.Success {
			// If no specific retry status codes, retry on any failure
			shouldRetryRequest = true
		}

		if !shouldRetryRequest {
			break // Don't retry if status code doesn't match retry criteria
		}

		// Calculate delay for this attempt
		delay := calculateDelay(attempt, test.RetryDelay, test.RetryBackoff)

		// Wait before retrying
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return lastResult, ctx.Err()
		}
	}

	return lastResult, lastErr
}

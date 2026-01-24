package executor

import (
	"context"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		retryOnStatus  []int
		expectedResult bool
	}{
		{
			name:           "Empty retry list returns false",
			statusCode:     500,
			retryOnStatus:  []int{},
			expectedResult: false,
		},
		{
			name:           "Status in retry list returns true",
			statusCode:     503,
			retryOnStatus:  []int{500, 502, 503, 504},
			expectedResult: true,
		},
		{
			name:           "Status not in retry list returns false",
			statusCode:     404,
			retryOnStatus:  []int{500, 502, 503, 504},
			expectedResult: false,
		},
		{
			name:           "Single status match",
			statusCode:     500,
			retryOnStatus:  []int{500},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRetry(tt.statusCode, tt.retryOnStatus)
			if result != tt.expectedResult {
				t.Errorf("shouldRetry(%d, %v) = %v, want %v",
					tt.statusCode, tt.retryOnStatus, result, tt.expectedResult)
			}
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	tests := []struct {
		name         string
		attempt      int
		initialDelay time.Duration
		backoffType  string
		expectedMin  time.Duration
		expectedMax  time.Duration
	}{
		{
			name:         "Exponential backoff attempt 0",
			attempt:      0,
			initialDelay: 1 * time.Second,
			backoffType:  "exponential",
			expectedMin:  1 * time.Second,
			expectedMax:  1 * time.Second,
		},
		{
			name:         "Exponential backoff attempt 1",
			attempt:      1,
			initialDelay: 1 * time.Second,
			backoffType:  "exponential",
			expectedMin:  2 * time.Second,
			expectedMax:  2 * time.Second,
		},
		{
			name:         "Exponential backoff attempt 2",
			attempt:      2,
			initialDelay: 1 * time.Second,
			backoffType:  "exponential",
			expectedMin:  4 * time.Second,
			expectedMax:  4 * time.Second,
		},
		{
			name:         "Linear backoff attempt 0",
			attempt:      0,
			initialDelay: 1 * time.Second,
			backoffType:  "linear",
			expectedMin:  1 * time.Second,
			expectedMax:  1 * time.Second,
		},
		{
			name:         "Linear backoff attempt 1",
			attempt:      1,
			initialDelay: 1 * time.Second,
			backoffType:  "linear",
			expectedMin:  2 * time.Second,
			expectedMax:  2 * time.Second,
		},
		{
			name:         "Linear backoff attempt 2",
			attempt:      2,
			initialDelay: 1 * time.Second,
			backoffType:  "linear",
			expectedMin:  3 * time.Second,
			expectedMax:  3 * time.Second,
		},
		{
			name:         "Zero initial delay defaults to 1s",
			attempt:      0,
			initialDelay: 0,
			backoffType:  "exponential",
			expectedMin:  1 * time.Second,
			expectedMax:  1 * time.Second,
		},
		{
			name:         "Unknown backoff type defaults to exponential",
			attempt:      1,
			initialDelay: 1 * time.Second,
			backoffType:  "unknown",
			expectedMin:  2 * time.Second,
			expectedMax:  2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateDelay(tt.attempt, tt.initialDelay, tt.backoffType)
			if delay < tt.expectedMin || delay > tt.expectedMax {
				t.Errorf("calculateDelay(%d, %v, %s) = %v, want between %v and %v",
					tt.attempt, tt.initialDelay, tt.backoffType, delay, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestExecuteWithRetry_NoRetries(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name:    "No retry test",
		Curl:    "curl https://httpbin.org/status/200",
		Retries: 0, // No retries
		Assertions: []models.Assertion{
			{Type: models.AssertionStatus, Value: "200"},
		},
	}

	ctx := context.Background()
	result, err := executor.ExecuteWithRetry(ctx, test)

	if err != nil {
		t.Fatalf("ExecuteWithRetry failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

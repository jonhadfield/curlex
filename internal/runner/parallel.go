package runner

import (
	"context"
	"sync"
	"time"

	"curlex/internal/models"
)

// RunParallel executes tests in parallel with controlled concurrency
func (r *Runner) RunParallel(ctx context.Context, suite *models.TestSuite, concurrency int, failFast bool) (*models.SuiteResult, error) {
	startTime := time.Now()

	// Default concurrency to 10 if not specified
	if concurrency <= 0 {
		concurrency = 10
	}

	// Create channels for work distribution
	jobs := make(chan models.Test, len(suite.Tests))
	results := make(chan models.TestResult, len(suite.Tests))

	// Context for cancellation (fail-fast)
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Worker pool
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for test := range jobs {
				// Check if context is cancelled (fail-fast)
				select {
				case <-runCtx.Done():
					return
				default:
				}

				// Execute the test
				result, err := r.executor.ExecuteWithRetry(runCtx, test)
				if err != nil {
					// Create error result
					result = &models.TestResult{
						Test:    test,
						Success: false,
						Error:   err,
					}
				}

				// Run assertions if no error occurred
				if result.Error == nil {
					failures := r.engine.Validate(result, test.Assertions)
					result.Failures = failures
					result.Success = len(failures) == 0
				}

				// Log request/response if logging is enabled
				if r.logger != nil {
					_ = r.logger.LogTest(*result, result.PreparedRequest)
				}

				// Send result
				select {
				case results <- *result:
					// Update progress if enabled
					if r.progress != nil {
						r.progress.Increment()
					}
					// If fail-fast is enabled and test failed, cancel context
					if failFast && !result.Success {
						cancel()
					}
				case <-runCtx.Done():
					return
				}
			}
		}()
	}

	// Send all tests to workers
	go func() {
		for _, test := range suite.Tests {
			select {
			case jobs <- test:
			case <-runCtx.Done():
				close(jobs)
				return
			}
		}
		close(jobs)
	}()

	// Collect results
	var testResults []models.TestResult
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		testResults = append(testResults, result)
	}

	endTime := time.Now()

	// Calculate stats
	passed := 0
	failed := 0
	for _, result := range testResults {
		if result.Success {
			passed++
		} else {
			failed++
		}
	}

	suiteResult := &models.SuiteResult{
		Results:     testResults,
		TotalTests:  len(testResults),
		PassedTests: passed,
		FailedTests: failed,
		TotalTime:   endTime.Sub(startTime),
		StartTime:   startTime,
		EndTime:     endTime,
	}

	return suiteResult, nil
}

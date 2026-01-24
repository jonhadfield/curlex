package runner

import (
	"context"
	"fmt"
	"os"
	"time"

	"curlex/internal/assertion"
	"curlex/internal/executor"
	"curlex/internal/models"
	"curlex/internal/output"
)

// Runner executes test suites
type Runner struct {
	executor *executor.Executor
	engine   *assertion.Engine
	logger   *output.RequestLogger
	progress *output.Progress
}

// NewRunner creates a new test runner
func NewRunner(timeout time.Duration, logDir string) *Runner {
	return &Runner{
		executor: executor.NewExecutor(timeout),
		engine:   assertion.NewEngine(),
		logger:   output.NewRequestLogger(logDir),
	}
}

// SetProgress sets the progress indicator for this runner
func (r *Runner) SetProgress(progress *output.Progress) {
	r.progress = progress
}

// Run executes all tests in the suite sequentially
func (r *Runner) Run(ctx context.Context, suite *models.TestSuite) (*models.SuiteResult, error) {
	startTime := time.Now()
	var results []models.TestResult

	for _, test := range suite.Tests {
		// Execute the test (with retry if configured)
		result, err := r.executor.ExecuteWithRetry(ctx, test)
		if err != nil {
			return nil, err
		}

		// Run assertions if no error occurred
		if result.Error == nil {
			failures := r.engine.Validate(result, test.Assertions)
			result.Failures = failures
			result.Success = len(failures) == 0
		}

		// Log request/response if logging is enabled
		if r.logger != nil {
			if err := r.logger.LogTest(*result, result.PreparedRequest); err != nil {
				// Don't fail the test, but warn the user about logging issues
				fmt.Fprintf(os.Stderr, "Warning: failed to write log file: %v\n", err)
			}
		}

		results = append(results, *result)

		// Update progress if enabled
		if r.progress != nil {
			r.progress.Increment()
		}
	}

	endTime := time.Now()

	// Calculate stats
	passed := 0
	failed := 0
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
		}
	}

	suiteResult := &models.SuiteResult{
		Results:     results,
		TotalTests:  len(results),
		PassedTests: passed,
		FailedTests: failed,
		TotalTime:   endTime.Sub(startTime),
		StartTime:   startTime,
		EndTime:     endTime,
	}

	return suiteResult, nil
}

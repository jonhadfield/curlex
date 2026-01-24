package models

import (
	"fmt"
	"net/http"
	"time"
)

// TestResult represents the result of executing a single test
type TestResult struct {
	Test            Test
	Success         bool
	StatusCode      int
	ResponseTime    time.Duration
	ResponseBody    string
	Headers         http.Header
	Failures        []AssertionFailure
	Error           error
	PreparedRequest *PreparedRequest // Request details for logging
}

// AssertionFailure represents a failed assertion with details
type AssertionFailure struct {
	Type     AssertionType
	Expected string
	Actual   string
	Message  string
}

// String returns a human-readable representation of the failure
func (f AssertionFailure) String() string {
	if f.Message != "" {
		return f.Message
	}
	return fmt.Sprintf("expected %s, got %s", f.Expected, f.Actual)
}

// SuiteResult represents the overall test suite execution results
type SuiteResult struct {
	Results      []TestResult
	TotalTests   int
	PassedTests  int
	FailedTests  int
	TotalTime    time.Duration
	StartTime    time.Time
	EndTime      time.Time
}

// HasFailures returns true if any test failed
func (sr SuiteResult) HasFailures() bool {
	return sr.FailedTests > 0
}

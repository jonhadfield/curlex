package parser

import (
	"curlex/internal/models"
)

// MergeDefaults applies default configuration to a test
// Test-level settings override defaults
func MergeDefaults(test *models.Test, defaults models.DefaultConfig) {
	// Apply timeout if not set on test
	if test.Timeout == 0 && defaults.Timeout > 0 {
		test.Timeout = defaults.Timeout
	}

	// Apply retries if not set on test
	if test.Retries == 0 && defaults.Retries > 0 {
		test.Retries = defaults.Retries
	}

	// Apply retry_delay if not set on test
	if test.RetryDelay == 0 && defaults.RetryDelay > 0 {
		test.RetryDelay = defaults.RetryDelay
	}

	// Apply retry_backoff if not set on test
	if test.RetryBackoff == "" && defaults.RetryBackoff != "" {
		test.RetryBackoff = defaults.RetryBackoff
	}

	// Apply retry_on_status if not set on test
	if len(test.RetryOnStatus) == 0 && len(defaults.RetryOnStatus) > 0 {
		test.RetryOnStatus = make([]int, len(defaults.RetryOnStatus))
		copy(test.RetryOnStatus, defaults.RetryOnStatus)
	}

	// Apply max_redirects if not set on test
	if test.MaxRedirects == nil && defaults.MaxRedirects != nil {
		redirects := *defaults.MaxRedirects
		test.MaxRedirects = &redirects
	}

	// Merge headers for structured requests
	if test.Request != nil && len(defaults.Headers) > 0 {
		mergeHeaders(test.Request, defaults.Headers)
	}
}

// mergeHeaders merges default headers into request headers
// Request headers override default headers
func mergeHeaders(request *models.StructuredRequest, defaultHeaders map[string]string) {
	if request.Headers == nil {
		request.Headers = make(map[string]string)
	}

	// Add default headers that aren't already set
	for key, value := range defaultHeaders {
		if _, exists := request.Headers[key]; !exists {
			request.Headers[key] = value
		}
	}
}

// ApplyDefaults applies defaults to all tests in a suite
func ApplyDefaults(suite *models.TestSuite) {
	for i := range suite.Tests {
		MergeDefaults(&suite.Tests[i], suite.Defaults)
	}
}

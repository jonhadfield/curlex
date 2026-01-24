package parser

import (
	"time"

	"curlex/internal/models"
)

// Validation constants for sane defaults
const (
	maxTimeout    = 10 * time.Minute
	maxRetries    = 100
	maxRedirects  = 1000
)

// MergeDefaults applies default configuration to a test
// Test-level settings override defaults
func MergeDefaults(test *models.Test, defaults models.DefaultConfig) {
	// Apply timeout if not set on test (with validation)
	if test.Timeout == 0 && defaults.Timeout > 0 {
		// Cap timeout at reasonable maximum
		if defaults.Timeout > maxTimeout {
			test.Timeout = maxTimeout
		} else {
			test.Timeout = defaults.Timeout
		}
	}

	// Apply retries if not set on test (with validation)
	if test.Retries == 0 && defaults.Retries > 0 {
		// Cap retries at reasonable maximum
		if defaults.Retries > maxRetries {
			test.Retries = maxRetries
		} else {
			test.Retries = defaults.Retries
		}
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

	// Apply max_redirects if not set on test (with validation)
	if test.MaxRedirects == nil && defaults.MaxRedirects != nil {
		redirects := *defaults.MaxRedirects
		// Validate redirect count is sane (allow -1 for unlimited)
		if redirects > maxRedirects && redirects != -1 {
			redirects = maxRedirects
		}
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

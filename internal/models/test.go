package models

import "time"

// TestSuite represents a collection of tests defined in a YAML file
type TestSuite struct {
	Version   string            `yaml:"version"`
	Variables map[string]string `yaml:"variables"`
	Defaults  DefaultConfig     `yaml:"defaults"`
	Tests     []Test            `yaml:"tests"`
}

// DefaultConfig holds default configuration for all tests
type DefaultConfig struct {
	Timeout       time.Duration     `yaml:"timeout"`
	Retries       int               `yaml:"retries"`
	RetryDelay    time.Duration     `yaml:"retry_delay,omitempty"`     // Delay between retries
	RetryBackoff  string            `yaml:"retry_backoff,omitempty"`   // "exponential" or "linear"
	RetryOnStatus []int             `yaml:"retry_on_status,omitempty"` // Status codes to retry on
	Headers       map[string]string `yaml:"headers"`
	MaxRedirects  *int              `yaml:"max_redirects,omitempty"` // nil = default (10), 0 = no redirects, -1 = unlimited
}

// Test represents a single HTTP test case
type Test struct {
	Name          string             `yaml:"name"`
	Curl          string             `yaml:"curl,omitempty"`
	Request       *StructuredRequest `yaml:"request,omitempty"`
	Assertions    []Assertion        `yaml:"assertions"`
	Timeout       time.Duration      `yaml:"timeout,omitempty"`
	Retries       int                `yaml:"retries,omitempty"`
	RetryDelay    time.Duration      `yaml:"retry_delay,omitempty"`     // Delay between retries
	RetryBackoff  string             `yaml:"retry_backoff,omitempty"`   // "exponential" or "linear"
	RetryOnStatus []int              `yaml:"retry_on_status,omitempty"` // Status codes to retry on
	MaxRedirects  *int               `yaml:"max_redirects,omitempty"`   // nil = default (10), 0 = no redirects, -1 = unlimited
	Debug         bool               `yaml:"debug,omitempty"`           // Print response headers and body for debugging
}

// StructuredRequest represents an HTTP request in structured format
type StructuredRequest struct {
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
}

// PreparedRequest is the internal representation after parsing curl or structured request
type PreparedRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

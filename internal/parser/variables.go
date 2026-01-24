package parser

import (
	"fmt"
	"os"
	"regexp"

	"curlex/internal/models"
)

// Pre-compiled regex pattern for variable substitution
var variablePattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// VariableExpander handles variable substitution in test suites
type VariableExpander struct {
	variables map[string]string
}

// NewVariableExpander creates a new variable expander
func NewVariableExpander() *VariableExpander {
	return &VariableExpander{
		variables: make(map[string]string),
	}
}

// ExpandVariables substitutes ${VAR_NAME} references in the test suite
func (ve *VariableExpander) ExpandVariables(suite *models.TestSuite) error {
	// Build variable map: test-level vars + environment vars
	ve.variables = make(map[string]string)

	// Add test-level variables first
	if suite.Variables != nil {
		for key, value := range suite.Variables {
			ve.variables[key] = value
		}
	}

	// Expand environment variables in test-level variables
	for key, value := range ve.variables {
		ve.variables[key] = ve.expandString(value)
	}

	// Expand variables in tests
	for i := range suite.Tests {
		if err := ve.expandTest(&suite.Tests[i]); err != nil {
			return fmt.Errorf("test %s: %w", suite.Tests[i].Name, err)
		}
	}

	return nil
}

// expandTest expands variables in a single test
func (ve *VariableExpander) expandTest(test *models.Test) error {
	// Expand curl command
	if test.Curl != "" {
		test.Curl = ve.expandString(test.Curl)
	}

	// Expand structured request
	if test.Request != nil {
		test.Request.URL = ve.expandString(test.Request.URL)
		test.Request.Body = ve.expandString(test.Request.Body)

		// Expand headers
		if test.Request.Headers != nil {
			expandedHeaders := make(map[string]string)
			for key, value := range test.Request.Headers {
				expandedKey := ve.expandString(key)
				expandedValue := ve.expandString(value)
				expandedHeaders[expandedKey] = expandedValue
			}
			test.Request.Headers = expandedHeaders
		}
	}

	// Expand assertions
	for i := range test.Assertions {
		test.Assertions[i].Value = ve.expandString(test.Assertions[i].Value)
	}

	return nil
}

// expandString replaces ${VAR_NAME} with variable values
func (ve *VariableExpander) expandString(s string) string {
	return variablePattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name (remove ${ and })
		varName := match[2 : len(match)-1]

		// Look up in test-level variables first
		if value, ok := ve.variables[varName]; ok {
			return value
		}

		// Fall back to environment variable
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// If not found, keep the original placeholder
		return match
	})
}

// GetVariables returns the current variable map (for debugging)
func (ve *VariableExpander) GetVariables() map[string]string {
	result := make(map[string]string)
	for k, v := range ve.variables {
		result[k] = v
	}
	return result
}

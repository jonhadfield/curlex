package assertion

import (
	"fmt"
	"strconv"
	"strings"

	"curlex/internal/models"
)

// HeaderValidator validates response header assertions
type HeaderValidator struct{}

// Validate checks if the response headers match the assertion
func (v *HeaderValidator) Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure {
	// Parse the assertion: "Header-Name operator value"
	// Examples: "Content-Type == 'application/json'", "Content-Type contains json"

	expr := strings.TrimSpace(assertion.Value)

	// Parse expression
	headerName, operator, expectedValue, err := v.parseExpression(expr)
	if err != nil {
		return &models.AssertionFailure{
			Type:    models.AssertionHeader,
			Message: fmt.Sprintf("invalid expression: %v", err),
		}
	}

	// Get actual header value (case-insensitive)
	actualValue := v.getHeader(result.Headers, headerName)

	// Check if header exists
	if actualValue == "" {
		return &models.AssertionFailure{
			Type:     models.AssertionHeader,
			Expected: fmt.Sprintf("header %q to exist", headerName),
			Actual:   "header not found",
			Message:  fmt.Sprintf("header %q not found in response", headerName),
		}
	}

	// Evaluate the condition
	if !v.evaluateCondition(actualValue, operator, expectedValue) {
		return &models.AssertionFailure{
			Type:     models.AssertionHeader,
			Expected: fmt.Sprintf("%s %s %s", headerName, operator, expectedValue),
			Actual:   fmt.Sprintf("%s = %s", headerName, actualValue),
			Message:  fmt.Sprintf("%s %s %s failed: got %s", headerName, operator, expectedValue, actualValue),
		}
	}

	return nil // Success
}

// parseExpression parses a header assertion expression
// Format: "Header-Name operator value"
// Returns: headerName, operator, value, error
func (v *HeaderValidator) parseExpression(expr string) (string, string, string, error) {
	// Operators in order (longest first)
	operators := []string{" contains ", " == ", " != ", " > ", " < ", " >= ", " <= "}

	for _, op := range operators {
		if idx := strings.Index(expr, op); idx != -1 {
			headerName := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+len(op):])
			// Remove quotes from value if present
			value = strings.Trim(value, `"'`)
			return headerName, strings.TrimSpace(op), value, nil
		}
	}

	return "", "", "", fmt.Errorf("no valid operator found in expression: %s", expr)
}

// getHeader retrieves a header value (case-insensitive)
func (v *HeaderValidator) getHeader(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) {
			if len(values) > 0 {
				return values[0] // Return first value
			}
		}
	}
	return ""
}

// evaluateCondition evaluates a comparison between actual and expected header values
func (v *HeaderValidator) evaluateCondition(actual, operator, expected string) bool {
	switch operator {
	case "==":
		return actual == expected
	case "!=":
		return actual != expected
	case "contains":
		return strings.Contains(actual, expected)
	case ">":
		// Try numeric comparison
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(expected, 64)
		if err1 == nil && err2 == nil {
			return actualNum > expectedNum
		}
		// Fall back to string comparison
		return actual > expected
	case "<":
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(expected, 64)
		if err1 == nil && err2 == nil {
			return actualNum < expectedNum
		}
		return actual < expected
	case ">=":
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(expected, 64)
		if err1 == nil && err2 == nil {
			return actualNum >= expectedNum
		}
		return actual >= expected
	case "<=":
		actualNum, err1 := strconv.ParseFloat(actual, 64)
		expectedNum, err2 := strconv.ParseFloat(expected, 64)
		if err1 == nil && err2 == nil {
			return actualNum <= expectedNum
		}
		return actual <= expected
	default:
		return false
	}
}

package assertion

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"curlex/internal/models"
)

// StatusValidator validates HTTP status code assertions
type StatusValidator struct{}

// Validate checks if the status code matches the assertion
func (v *StatusValidator) Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure {
	expected := strings.TrimSpace(assertion.Value)
	actual := result.StatusCode

	// Check if it's a simple exact match (e.g., "200")
	if expectedCode, err := strconv.Atoi(expected); err == nil {
		if actual == expectedCode {
			return nil // Success
		}
		return &models.AssertionFailure{
			Type:     models.AssertionStatus,
			Expected: expected,
			Actual:   strconv.Itoa(actual),
			Message:  fmt.Sprintf("expected status %s, got %d", expected, actual),
		}
	}

	// Check if it's an expression (e.g., ">= 200 && < 300")
	if v.isExpression(expected) {
		if v.evaluateExpression(expected, actual) {
			return nil // Success
		}
		return &models.AssertionFailure{
			Type:     models.AssertionStatus,
			Expected: expected,
			Actual:   strconv.Itoa(actual),
			Message:  fmt.Sprintf("status %d does not satisfy expression: %s", actual, expected),
		}
	}

	// Invalid assertion format
	return &models.AssertionFailure{
		Type:    models.AssertionStatus,
		Message: fmt.Sprintf("invalid status assertion format: %s", expected),
	}
}

// isExpression checks if the status assertion is an expression
func (v *StatusValidator) isExpression(s string) bool {
	operators := []string{">=", "<=", "!=", "==", ">", "<", "&&", "||"}
	for _, op := range operators {
		if strings.Contains(s, op) {
			return true
		}
	}
	return false
}

// evaluateExpression evaluates a status code expression
// Supports: ==, !=, >, <, >=, <=, &&, ||
func (v *StatusValidator) evaluateExpression(expr string, actual int) bool {
	// Replace 'status' variable with actual value
	expr = strings.ReplaceAll(expr, "status", strconv.Itoa(actual))

	// Handle compound expressions with && and ||
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		for _, part := range parts {
			if !v.evaluateSingleExpression(strings.TrimSpace(part), actual) {
				return false
			}
		}
		return true
	}

	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		for _, part := range parts {
			if v.evaluateSingleExpression(strings.TrimSpace(part), actual) {
				return true
			}
		}
		return false
	}

	// Single expression
	return v.evaluateSingleExpression(expr, actual)
}

// evaluateSingleExpression evaluates a single comparison
func (v *StatusValidator) evaluateSingleExpression(expr string, actual int) bool {
	// Pattern can be:
	// 1. "number operator number" e.g., "200 >= 200", "404 != 200"
	// 2. "operator number" e.g., ">= 200", "< 300" (actual is implicit left operand)

	// Try pattern: "operator number" (actual is left operand)
	// >= operator
	if match := v.extractComparison(expr, `^\s*>=\s*(\d+)`); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual >= right
	}

	// <= operator
	if match := v.extractComparison(expr, `^\s*<=\s*(\d+)`); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual <= right
	}

	// > operator
	if match := v.extractComparison(expr, `^\s*>\s*(\d+)`); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual > right
	}

	// < operator
	if match := v.extractComparison(expr, `^\s*<\s*(\d+)`); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual < right
	}

	// == operator
	if match := v.extractComparison(expr, `^\s*==\s*(\d+)`); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual == right
	}

	// != operator
	if match := v.extractComparison(expr, `^\s*!=\s*(\d+)`); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual != right
	}

	// Try pattern: "number operator number"
	// >= operator
	if match := v.extractComparison(expr, `(\d+)\s*>=\s*(\d+)`); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left >= right
	}

	// <= operator
	if match := v.extractComparison(expr, `(\d+)\s*<=\s*(\d+)`); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left <= right
	}

	// > operator
	if match := v.extractComparison(expr, `(\d+)\s*>\s*(\d+)`); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left > right
	}

	// < operator
	if match := v.extractComparison(expr, `(\d+)\s*<\s*(\d+)`); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left < right
	}

	// == operator
	if match := v.extractComparison(expr, `(\d+)\s*==\s*(\d+)`); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left == right
	}

	// != operator
	if match := v.extractComparison(expr, `(\d+)\s*!=\s*(\d+)`); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left != right
	}

	return false
}

// extractComparison extracts comparison operands from expression
func (v *StatusValidator) extractComparison(expr, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(expr)
	if len(matches) >= 2 {
		return matches
	}
	return nil
}

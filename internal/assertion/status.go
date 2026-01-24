package assertion

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"curlex/internal/models"
)

// Pre-compiled regex patterns for status code comparisons
var (
	// "operator number" patterns (actual is left operand)
	statusPatternGTE     = regexp.MustCompile(`^\s*>=\s*(\d+)`)
	statusPatternLTE     = regexp.MustCompile(`^\s*<=\s*(\d+)`)
	statusPatternGT      = regexp.MustCompile(`^\s*>\s*(\d+)`)
	statusPatternLT      = regexp.MustCompile(`^\s*<\s*(\d+)`)
	statusPatternEQ      = regexp.MustCompile(`^\s*==\s*(\d+)`)
	statusPatternNEQ     = regexp.MustCompile(`^\s*!=\s*(\d+)`)

	// "number operator number" patterns
	statusPatternNumGTE  = regexp.MustCompile(`(\d+)\s*>=\s*(\d+)`)
	statusPatternNumLTE  = regexp.MustCompile(`(\d+)\s*<=\s*(\d+)`)
	statusPatternNumGT   = regexp.MustCompile(`(\d+)\s*>\s*(\d+)`)
	statusPatternNumLT   = regexp.MustCompile(`(\d+)\s*<\s*(\d+)`)
	statusPatternNumEQ   = regexp.MustCompile(`(\d+)\s*==\s*(\d+)`)
	statusPatternNumNEQ  = regexp.MustCompile(`(\d+)\s*!=\s*(\d+)`)
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
	// Replace 'status' variable with actual value only if present
	if strings.Contains(expr, "status") {
		expr = strings.ReplaceAll(expr, "status", strconv.Itoa(actual))
	}

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
	if match := v.extractComparisonRe(expr, statusPatternGTE); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual >= right
	}

	// <= operator
	if match := v.extractComparisonRe(expr, statusPatternLTE); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual <= right
	}

	// > operator
	if match := v.extractComparisonRe(expr, statusPatternGT); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual > right
	}

	// < operator
	if match := v.extractComparisonRe(expr, statusPatternLT); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual < right
	}

	// == operator
	if match := v.extractComparisonRe(expr, statusPatternEQ); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual == right
	}

	// != operator
	if match := v.extractComparisonRe(expr, statusPatternNEQ); match != nil {
		right, _ := strconv.Atoi(match[1])
		return actual != right
	}

	// Try pattern: "number operator number"
	// >= operator
	if match := v.extractComparisonRe(expr, statusPatternNumGTE); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left >= right
	}

	// <= operator
	if match := v.extractComparisonRe(expr, statusPatternNumLTE); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left <= right
	}

	// > operator
	if match := v.extractComparisonRe(expr, statusPatternNumGT); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left > right
	}

	// < operator
	if match := v.extractComparisonRe(expr, statusPatternNumLT); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left < right
	}

	// == operator
	if match := v.extractComparisonRe(expr, statusPatternNumEQ); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left == right
	}

	// != operator
	if match := v.extractComparisonRe(expr, statusPatternNumNEQ); match != nil {
		left, _ := strconv.Atoi(match[1])
		right, _ := strconv.Atoi(match[2])
		return left != right
	}

	return false
}

// extractComparisonRe extracts comparison operands using pre-compiled regex
func (v *StatusValidator) extractComparisonRe(expr string, re *regexp.Regexp) []string {
	matches := re.FindStringSubmatch(expr)
	if len(matches) >= 2 {
		return matches
	}
	return nil
}

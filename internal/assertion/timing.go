package assertion

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"curlex/internal/models"
)

// ResponseTimeValidator validates response time assertions
type ResponseTimeValidator struct{}

// Validate checks if the response time meets the assertion
func (v *ResponseTimeValidator) Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure {
	// Parse the assertion: "< 500ms", "<= 2s"
	expr := strings.TrimSpace(assertion.Value)

	// Parse expression
	operator, duration, err := v.parseExpression(expr)
	if err != nil {
		return &models.AssertionFailure{
			Type:    models.AssertionResponseTime,
			Message: fmt.Sprintf("invalid expression: %v", err),
		}
	}

	actual := result.ResponseTime

	// Evaluate the condition
	if !v.evaluateCondition(actual, operator, duration) {
		return &models.AssertionFailure{
			Type:     models.AssertionResponseTime,
			Expected: fmt.Sprintf("%s %s", operator, duration),
			Actual:   actual.String(),
			Message:  fmt.Sprintf("response time %s does not satisfy %s %s", actual, operator, duration),
		}
	}

	return nil // Success
}

// parseExpression parses a response time expression
// Format: "operator duration" (e.g., "< 500ms", "<= 2s")
// Returns: operator, duration, error
func (v *ResponseTimeValidator) parseExpression(expr string) (string, time.Duration, error) {
	// Pattern: operator + duration
	// Examples: "< 500ms", "<= 2s", "> 100ms"

	// Extract operator
	operators := []string{"<=", ">=", "<", ">", "==", "!="}
	var operator string
	var durationStr string

	for _, op := range operators {
		if strings.HasPrefix(expr, op) {
			operator = op
			durationStr = strings.TrimSpace(expr[len(op):])
			break
		}
	}

	if operator == "" {
		return "", 0, fmt.Errorf("no valid operator found in expression: %s", expr)
	}

	// Parse duration
	duration, err := v.parseDuration(durationStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid duration %q: %w", durationStr, err)
	}

	return operator, duration, nil
}

// parseDuration parses duration strings like "500ms", "2s", "1000ms"
func (v *ResponseTimeValidator) parseDuration(s string) (time.Duration, error) {
	// Pattern: number + unit
	pattern := regexp.MustCompile(`^(\d+(?:\.\d+)?)(ms|s|m|h)$`)
	matches := pattern.FindStringSubmatch(s)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	unit := matches[2]

	switch unit {
	case "ms":
		return time.Duration(value * float64(time.Millisecond)), nil
	case "s":
		return time.Duration(value * float64(time.Second)), nil
	case "m":
		return time.Duration(value * float64(time.Minute)), nil
	case "h":
		return time.Duration(value * float64(time.Hour)), nil
	default:
		return 0, fmt.Errorf("unknown time unit: %s", unit)
	}
}

// evaluateCondition evaluates a comparison between actual and expected durations
func (v *ResponseTimeValidator) evaluateCondition(actual time.Duration, operator string, expected time.Duration) bool {
	switch operator {
	case "<":
		return actual < expected
	case "<=":
		return actual <= expected
	case ">":
		return actual > expected
	case ">=":
		return actual >= expected
	case "==":
		return actual == expected
	case "!=":
		return actual != expected
	default:
		return false
	}
}

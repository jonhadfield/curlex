package assertion

import (
	"fmt"
	"strconv"
	"strings"

	"curlex/internal/models"
	"github.com/tidwall/gjson"
)

// JSONPathValidator validates JSON path assertions
type JSONPathValidator struct{}

// Validate checks if the JSON path expression evaluates to true
func (v *JSONPathValidator) Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure {
	// Parse the assertion: ".path operator value"
	// Examples: ".data.id == 123", ".users[0].age > 18", ".active == true"

	expr := strings.TrimSpace(assertion.Value)

	// Split into path and condition
	path, operator, expectedValue, err := v.parseExpression(expr)
	if err != nil {
		return &models.AssertionFailure{
			Type:    models.AssertionJSONPath,
			Message: fmt.Sprintf("invalid expression: %v", err),
		}
	}

	// Remove leading dot for gjson (it doesn't use dots at the beginning)
	path = strings.TrimPrefix(path, ".")

	// Extract actual value from JSON using gjson
	jsonResult := gjson.Get(result.ResponseBody, path)

	// Check if path exists
	if !jsonResult.Exists() {
		return &models.AssertionFailure{
			Type:     models.AssertionJSONPath,
			Expected: fmt.Sprintf("path %q to exist", path),
			Actual:   "path does not exist",
			Message:  fmt.Sprintf("JSON path %q not found", path),
		}
	}

	// Evaluate the condition
	if !v.evaluateCondition(jsonResult, operator, expectedValue) {
		return &models.AssertionFailure{
			Type:     models.AssertionJSONPath,
			Expected: fmt.Sprintf("%s %s %s", path, operator, expectedValue),
			Actual:   fmt.Sprintf("%s = %v", path, jsonResult.Value()),
			Message:  fmt.Sprintf("%s %s %s failed: got %v", path, operator, expectedValue, jsonResult.Value()),
		}
	}

	return nil // Success
}

// parseExpression parses a JSON path expression
// Format: ".path operator value"
// Returns: path, operator, value, error
func (v *JSONPathValidator) parseExpression(expr string) (string, string, string, error) {
	// Operators in order of precedence (longest first to match correctly)
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}

	for _, op := range operators {
		if idx := strings.Index(expr, " "+op+" "); idx != -1 {
			path := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+len(op)+2:])
			return path, op, value, nil
		}
	}

	return "", "", "", fmt.Errorf("no valid operator found in expression: %s", expr)
}

// evaluateCondition evaluates a comparison between gjson result and expected value
func (v *JSONPathValidator) evaluateCondition(actual gjson.Result, operator, expected string) bool {
	// Handle different types
	switch actual.Type {
	case gjson.String:
		return v.evaluateString(actual.String(), operator, expected)
	case gjson.Number:
		return v.evaluateNumber(actual.Float(), operator, expected)
	case gjson.True, gjson.False:
		return v.evaluateBool(actual.Bool(), operator, expected)
	case gjson.Null:
		return v.evaluateNull(operator, expected)
	default:
		// For complex types, convert to string
		return v.evaluateString(actual.String(), operator, expected)
	}
}

// evaluateString compares string values
func (v *JSONPathValidator) evaluateString(actual, operator, expected string) bool {
	// Remove quotes from expected if present
	expected = strings.Trim(expected, `"'`)

	switch operator {
	case "==":
		return actual == expected
	case "!=":
		return actual != expected
	case ">":
		return actual > expected
	case "<":
		return actual < expected
	case ">=":
		return actual >= expected
	case "<=":
		return actual <= expected
	default:
		return false
	}
}

// evaluateNumber compares numeric values
func (v *JSONPathValidator) evaluateNumber(actual float64, operator, expected string) bool {
	expectedNum, err := strconv.ParseFloat(expected, 64)
	if err != nil {
		return false
	}

	switch operator {
	case "==":
		return actual == expectedNum
	case "!=":
		return actual != expectedNum
	case ">":
		return actual > expectedNum
	case "<":
		return actual < expectedNum
	case ">=":
		return actual >= expectedNum
	case "<=":
		return actual <= expectedNum
	default:
		return false
	}
}

// evaluateBool compares boolean values
func (v *JSONPathValidator) evaluateBool(actual bool, operator, expected string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))
	expectedBool := expected == "true"

	switch operator {
	case "==":
		return actual == expectedBool
	case "!=":
		return actual != expectedBool
	default:
		return false
	}
}

// evaluateNull checks null conditions
func (v *JSONPathValidator) evaluateNull(operator, expected string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))

	switch operator {
	case "==":
		return expected == "null"
	case "!=":
		return expected != "null"
	default:
		return false
	}
}

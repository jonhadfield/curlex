package assertion

import (
	"fmt"
	"strings"

	"curlex/internal/models"
)

// BodyValidator validates response body assertions
type BodyValidator struct{}

// Validate checks if the response body matches the assertion
func (v *BodyValidator) Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure {
	expected := assertion.Value
	actual := result.ResponseBody

	if actual == expected {
		return nil // Success
	}

	return &models.AssertionFailure{
		Type:     models.AssertionBody,
		Expected: expected,
		Actual:   v.truncate(actual, 100),
		Message:  "body mismatch: expected exact match",
	}
}

// truncate limits string length for display
func (v *BodyValidator) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// BodyContainsValidator validates body contains assertions
type BodyContainsValidator struct{}

// Validate checks if the response body contains the expected substring
func (v *BodyContainsValidator) Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure {
	substring := assertion.Value
	actual := result.ResponseBody

	if strings.Contains(actual, substring) {
		return nil // Success
	}

	return &models.AssertionFailure{
		Type:     models.AssertionBodyContains,
		Expected: fmt.Sprintf("body to contain: %q", substring),
		Actual:   v.truncate(actual, 100),
		Message:  fmt.Sprintf("body does not contain %q", substring),
	}
}

// truncate limits string length for display
func (v *BodyContainsValidator) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

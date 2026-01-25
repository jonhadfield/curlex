package assertion

import (
	"net/http"
	"testing"

	"curlex/internal/models"
)

func TestHeaderValidator_ExactMatch(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "Content-Type == 'application/json'",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestHeaderValidator_CaseInsensitive(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	// Test lowercase header name
	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "content-type == 'application/json'",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure (case insensitive), got: %v", failure.Message)
	}
}

func TestHeaderValidator_Contains(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{
			"Content-Type": []string{"application/json; charset=utf-8"},
		},
	}

	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "Content-Type contains json",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestHeaderValidator_NotEqual(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "Content-Type != 'text/html'",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestHeaderValidator_NumericComparison(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{
			"X-RateLimit-Remaining": []string{"100"},
		},
	}

	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "X-RateLimit-Remaining > 50",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestHeaderValidator_HeaderNotFound(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{},
	}

	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "Missing-Header == 'value'",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Fatal("Expected failure for missing header")
	}

	if failure.Actual != "header not found" {
		t.Errorf("Expected 'header not found', got: %s", failure.Actual)
	}
}

func TestHeaderValidator_InvalidExpression(t *testing.T) {
	validator := &HeaderValidator{}

	result := &models.TestResult{
		Headers: http.Header{},
	}

	assertion := models.Assertion{
		Type:  models.AssertionHeader,
		Value: "Invalid Expression Without Operator",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure for invalid expression")
	}
}

func TestHeaderValidator_AllOperators(t *testing.T) {
	tests := []struct {
		name       string
		headerVal  string
		expression string
		shouldPass bool
	}{
		{"GreaterThan", "100", "X-Value > 50", true},
		{"LessThan", "50", "X-Value < 100", true},
		{"GreaterOrEqual", "100", "X-Value >= 100", true},
		{"LessOrEqual", "50", "X-Value <= 50", true},
		{"Equal", "test", "X-Value == test", true},
		{"NotEqual", "test", "X-Value != other", true},
		{"Contains", "application/json", "X-Value contains json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &HeaderValidator{}

			result := &models.TestResult{
				Headers: http.Header{
					"X-Value": []string{tt.headerVal},
				},
			}

			assertion := models.Assertion{
				Type:  models.AssertionHeader,
				Value: tt.expression,
			}

			failure := validator.Validate(result, assertion)
			if tt.shouldPass && failure != nil {
				t.Errorf("Expected pass, got failure: %v", failure.Message)
			}
			if !tt.shouldPass && failure == nil {
				t.Error("Expected failure, got pass")
			}
		})
	}
}

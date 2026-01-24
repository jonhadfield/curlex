package assertion

import (
	"testing"

	"curlex/internal/models"
)

func TestStatusValidator_Simple(t *testing.T) {
	validator := &StatusValidator{}

	tests := []struct {
		name       string
		expected   string
		actualCode int
		shouldPass bool
	}{
		{"exact match pass", "200", 200, true},
		{"exact match fail", "200", 404, false},
		{"range expression >=", ">= 200", 201, true},
		{"range expression <", "< 300", 201, true},
		{"compound expression", ">= 200 && < 300", 201, true},
		{"compound expression fail", ">= 200 && < 300", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &models.TestResult{StatusCode: tt.actualCode}
			assertion := models.Assertion{
				Type:  models.AssertionStatus,
				Value: tt.expected,
			}

			failure := validator.Validate(result, assertion)

			if tt.shouldPass && failure != nil {
				t.Errorf("Expected to pass, but failed: %v", failure)
			}
			if !tt.shouldPass && failure == nil {
				t.Errorf("Expected to fail, but passed")
			}
		})
	}
}

func TestStatusValidator_Expression(t *testing.T) {
	validator := &StatusValidator{}

	// Test expression evaluation directly
	tests := []struct {
		expr     string
		actual   int
		expected bool
	}{
		{">= 200", 201, true},
		{">= 200", 199, false},
		{"< 300", 201, true},
		{"< 300", 300, false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := validator.evaluateExpression(tt.expr, tt.actual)
			if result != tt.expected {
				t.Errorf("evaluateExpression(%q, %d) = %v, want %v", tt.expr, tt.actual, result, tt.expected)
			}
		})
	}
}

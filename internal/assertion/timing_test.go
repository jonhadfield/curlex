package assertion

import (
	"testing"
	"time"

	"curlex/internal/models"
)

func TestResponseTimeValidator_LessThan(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 300 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "< 500ms",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestResponseTimeValidator_GreaterThan(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 600 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "> 500ms",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestResponseTimeValidator_LessThanOrEqual(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 500 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "<= 500ms",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestResponseTimeValidator_GreaterThanOrEqual(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 500 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: ">= 500ms",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestResponseTimeValidator_Seconds(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 1500 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "< 2s",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestResponseTimeValidator_Minutes(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 30 * time.Second,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "< 1m",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestResponseTimeValidator_Failure(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 600 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "< 500ms",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure, got none")
	}

	if failure.Type != models.AssertionResponseTime {
		t.Errorf("Expected failure type %v, got %v", models.AssertionResponseTime, failure.Type)
	}
}

func TestResponseTimeValidator_InvalidExpression(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 500 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "invalid expression",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure for invalid expression")
	}
}

func TestResponseTimeValidator_InvalidDuration(t *testing.T) {
	validator := &ResponseTimeValidator{}

	result := &models.TestResult{
		ResponseTime: 500 * time.Millisecond,
	}

	assertion := models.Assertion{
		Type:  models.AssertionResponseTime,
		Value: "< invalid",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure for invalid duration")
	}
}

func TestResponseTimeValidator_AllOperators(t *testing.T) {
	tests := []struct {
		name        string
		responseTime time.Duration
		expression  string
		shouldPass  bool
	}{
		{"LessThan", 400 * time.Millisecond, "< 500ms", true},
		{"GreaterThan", 600 * time.Millisecond, "> 500ms", true},
		{"LessThanOrEqual", 500 * time.Millisecond, "<= 500ms", true},
		{"GreaterThanOrEqual", 500 * time.Millisecond, ">= 500ms", true},
		{"LessThanFail", 600 * time.Millisecond, "< 500ms", false},
		{"GreaterThanFail", 400 * time.Millisecond, "> 500ms", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := &ResponseTimeValidator{}

			result := &models.TestResult{
				ResponseTime: tt.responseTime,
			}

			assertion := models.Assertion{
				Type:  models.AssertionResponseTime,
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

package assertion

import (
	"testing"

	"curlex/internal/models"
)

func TestBodyValidator_ExactMatch(t *testing.T) {
	validator := &BodyValidator{}

	result := &models.TestResult{
		ResponseBody: `{"status":"ok"}`,
	}

	// Exact match - should pass
	assertion := models.Assertion{
		Type:  models.AssertionBody,
		Value: `{"status":"ok"}`,
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestBodyValidator_ExactMismatch(t *testing.T) {
	validator := &BodyValidator{}

	result := &models.TestResult{
		ResponseBody: `{"status":"ok"}`,
	}

	// Exact mismatch - should fail
	assertion := models.Assertion{
		Type:  models.AssertionBody,
		Value: `{"status":"error"}`,
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure, got none")
	}

	if failure.Type != models.AssertionBody {
		t.Errorf("Expected failure type %v, got %v", models.AssertionBody, failure.Type)
	}
}

func TestBodyValidator_LongBodyTruncation(t *testing.T) {
	validator := &BodyValidator{}

	// Create a very long body
	longBody := ""
	for i := 0; i < 200; i++ {
		longBody += "a"
	}

	result := &models.TestResult{
		ResponseBody: longBody,
	}

	assertion := models.Assertion{
		Type:  models.AssertionBody,
		Value: "different",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Fatal("Expected failure")
	}

	// Check that actual value is truncated
	if len(failure.Actual) > 104 { // "..." + 100 chars + "..."
		t.Errorf("Expected truncated actual value, got length %d", len(failure.Actual))
	}
}

func TestBodyContainsValidator_Success(t *testing.T) {
	validator := &BodyContainsValidator{}

	result := &models.TestResult{
		ResponseBody: `{"status":"ok","message":"success"}`,
	}

	// Contains match - should pass
	assertion := models.Assertion{
		Type:  models.AssertionBodyContains,
		Value: "success",
	}

	failure := validator.Validate(result, assertion)
	if failure != nil {
		t.Errorf("Expected no failure, got: %v", failure.Message)
	}
}

func TestBodyContainsValidator_Failure(t *testing.T) {
	validator := &BodyContainsValidator{}

	result := &models.TestResult{
		ResponseBody: `{"status":"ok"}`,
	}

	// Does not contain - should fail
	assertion := models.Assertion{
		Type:  models.AssertionBodyContains,
		Value: "error",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure, got none")
	}

	if failure.Type != models.AssertionBodyContains {
		t.Errorf("Expected failure type %v, got %v", models.AssertionBodyContains, failure.Type)
	}
}

func TestBodyContainsValidator_EmptyBody(t *testing.T) {
	validator := &BodyContainsValidator{}

	result := &models.TestResult{
		ResponseBody: "",
	}

	assertion := models.Assertion{
		Type:  models.AssertionBodyContains,
		Value: "anything",
	}

	failure := validator.Validate(result, assertion)
	if failure == nil {
		t.Error("Expected failure for empty body")
	}
}

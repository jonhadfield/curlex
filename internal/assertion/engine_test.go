package assertion

import (
	"net/http"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestEngine_NewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("Expected engine, got nil")
	}
}

func TestEngine_StatusAssertion(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		StatusCode: 200,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionStatus, Value: "200"},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_BodyAssertion(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		ResponseBody: `{"status":"ok"}`,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionBody, Value: `{"status":"ok"}`},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_BodyContainsAssertion(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		ResponseBody: `{"status":"ok","message":"success"}`,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionBodyContains, Value: "success"},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_JSONPathAssertion(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		ResponseBody: `{"id":123,"name":"test"}`,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionJSONPath, Value: "id == 123"},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_HeaderAssertion(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	assertions := []models.Assertion{
		{Type: models.AssertionHeader, Value: "Content-Type == 'application/json'"},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_ResponseTimeAssertion(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		ResponseTime: 300 * time.Millisecond,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionResponseTime, Value: "< 500ms"},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_MultipleAssertions(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		StatusCode:   200,
		ResponseBody: `{"id":123,"name":"test"}`,
		ResponseTime: 300 * time.Millisecond,
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	assertions := []models.Assertion{
		{Type: models.AssertionStatus, Value: "200"},
		{Type: models.AssertionBodyContains, Value: "test"},
		{Type: models.AssertionJSONPath, Value: "id == 123"},
		{Type: models.AssertionHeader, Value: "Content-Type contains json"},
		{Type: models.AssertionResponseTime, Value: "< 500ms"},
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(failures))
	}
}

func TestEngine_MixedFailures(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		StatusCode:   404,
		ResponseBody: `{"error":"not found"}`,
		ResponseTime: 600 * time.Millisecond,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionStatus, Value: "200"},           // Fail
		{Type: models.AssertionBodyContains, Value: "error"},   // Pass
		{Type: models.AssertionResponseTime, Value: "< 500ms"}, // Fail
	}

	failures := engine.Validate(result, assertions)
	if len(failures) != 2 {
		t.Errorf("Expected 2 failures, got %d", len(failures))
	}
}

func TestEngine_UnknownAssertionType(t *testing.T) {
	engine := NewEngine()

	result := &models.TestResult{
		StatusCode: 200,
	}

	assertions := []models.Assertion{
		{Type: "unknown", Value: "test"},
	}

	failures := engine.Validate(result, assertions)
	// Unknown assertion types should not panic
	if failures == nil {
		t.Error("Expected failures slice to be initialized")
	}
}

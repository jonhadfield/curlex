package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"curlex/internal/assertion"
	"curlex/internal/executor"
	"curlex/internal/models"
	"curlex/internal/parser"
)

// BenchmarkHTTPRequest benchmarks basic HTTP request execution
func BenchmarkHTTPRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	exec := executor.NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Benchmark Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    server.URL,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exec.Execute(context.Background(), test)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkYAMLParsing benchmarks YAML parsing performance
func BenchmarkYAMLParsing(b *testing.B) {
	yamlContent := `version: "1.0"
tests:
  - name: "Test 1"
    request:
      method: GET
      url: "https://example.com"
    assertions:
      - status: 200
      - json_path: ".id == 1"
      - header: "Content-Type contains json"
`
	// Create temp file once
	tmpDir := b.TempDir()
	tmpFile := tmpDir + "/test.yaml"
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		b.Fatal(err)
	}

	p := parser.NewYAMLParser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Parse(tmpFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAssertionValidation benchmarks assertion validation
func BenchmarkAssertionValidation(b *testing.B) {
	result := &models.TestResult{
		StatusCode:   200,
		ResponseBody: `{"id": 123, "name": "test", "active": true}`,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		ResponseTime: 100 * time.Millisecond,
	}

	assertions := []models.Assertion{
		{Type: models.AssertionStatus, Value: "200"},
		{Type: models.AssertionBodyContains, Value: "test"},
		{Type: models.AssertionJSONPath, Value: ".id == 123"},
		{Type: models.AssertionHeader, Value: "Content-Type contains json"},
		{Type: models.AssertionResponseTime, Value: "< 1s"},
	}

	validators := map[models.AssertionType]assertion.Validator{
		models.AssertionStatus:       &assertion.StatusValidator{},
		models.AssertionBodyContains: &assertion.BodyValidator{},
		models.AssertionJSONPath:     &assertion.JSONPathValidator{},
		models.AssertionHeader:       &assertion.HeaderValidator{},
		models.AssertionResponseTime: &assertion.ResponseTimeValidator{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, assertion := range assertions {
			validator := validators[assertion.Type]
			if validator != nil {
				_ = validator.Validate(result, assertion)
			}
		}
	}
}

// BenchmarkJSONPathAssertion benchmarks complex JSON path assertions
func BenchmarkJSONPathAssertion(b *testing.B) {
	result := &models.TestResult{
		ResponseBody: `{
			"users": [
				{"id": 1, "name": "Alice", "age": 30, "active": true},
				{"id": 2, "name": "Bob", "age": 25, "active": false},
				{"id": 3, "name": "Charlie", "age": 35, "active": true}
			],
			"total": 3,
			"page": 1
		}`,
	}

	validator := &assertion.JSONPathValidator{}
	assertion := models.Assertion{
		Type:  models.AssertionJSONPath,
		Value: ".users[0].name == 'Alice'",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(result, assertion)
	}
}

// BenchmarkStatusRangeAssertion benchmarks status code range assertions
func BenchmarkStatusRangeAssertion(b *testing.B) {
	result := &models.TestResult{
		StatusCode: 200,
	}

	validator := &assertion.StatusValidator{}
	assertion := models.Assertion{
		Type:  models.AssertionStatus,
		Value: ">= 200 && < 300",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(result, assertion)
	}
}

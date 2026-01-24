package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"curlex/internal/models"
)

func BenchmarkSequentialExecution(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "Test 1", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 2", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 3", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 4", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 5", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
		},
	}

	runner := NewRunner(30*time.Second, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.Run(context.Background(), suite)
	}
}

func BenchmarkParallelExecution(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "Test 1", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 2", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 3", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 4", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
			{Name: "Test 5", Request: &models.StructuredRequest{Method: "GET", URL: server.URL}},
		},
	}

	runner := NewRunner(30*time.Second, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.RunParallel(context.Background(), suite, 10, false)
	}
}

func BenchmarkAssertionEngine(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":123,"name":"test","active":true}`))
	}))
	defer server.Close()

	suite := &models.TestSuite{
		Tests: []models.Test{
			{
				Name:    "Complex Assertions",
				Request: &models.StructuredRequest{Method: "GET", URL: server.URL},
				Assertions: []models.Assertion{
					{Type: models.AssertionStatus, Value: "200"},
					{Type: models.AssertionBodyContains, Value: "test"},
					{Type: models.AssertionJSONPath, Value: "id == 123"},
					{Type: models.AssertionJSONPath, Value: "name == 'test'"},
					{Type: models.AssertionHeader, Value: "Content-Type contains json"},
					{Type: models.AssertionResponseTime, Value: "< 5s"},
				},
			},
		},
	}

	runner := NewRunner(30*time.Second, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.Run(context.Background(), suite)
	}
}

func BenchmarkJSONPathAssertion(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}],"total":2}`))
	}))
	defer server.Close()

	suite := &models.TestSuite{
		Tests: []models.Test{
			{
				Name:    "JSON Path Tests",
				Request: &models.StructuredRequest{Method: "GET", URL: server.URL},
				Assertions: []models.Assertion{
					{Type: models.AssertionJSONPath, Value: "total == 2"},
					{Type: models.AssertionJSONPath, Value: "users.#.id == [1,2]"},
					{Type: models.AssertionJSONPath, Value: "users.0.name == 'Alice'"},
				},
			},
		},
	}

	runner := NewRunner(30*time.Second, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = runner.Run(context.Background(), suite)
	}
}

func BenchmarkRetryLogic(b *testing.B) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts%3 == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	suite := &models.TestSuite{
		Tests: []models.Test{
			{
				Name:          "Retry Test",
				Request:       &models.StructuredRequest{Method: "GET", URL: server.URL},
				Retries:       2,
				RetryDelay:    10 * time.Millisecond,
				RetryOnStatus: []int{500, 502, 503},
			},
		},
	}

	runner := NewRunner(30*time.Second, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts = 0
		_, _ = runner.Run(context.Background(), suite)
	}
}

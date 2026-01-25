package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestRunner_Integration_Sequential(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		case "/redirect":
			http.Redirect(w, r, "/success", http.StatusFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Create test suite
	suite := &models.TestSuite{
		Tests: []models.Test{
			{
				Name: "Success test",
				Request: &models.StructuredRequest{
					Method: "GET",
					URL:    server.URL + "/success",
				},
				Assertions: []models.Assertion{
					{Type: models.AssertionStatus, Value: "200"},
					{Type: models.AssertionBodyContains, Value: "ok"},
				},
			},
			{
				Name: "Not found test",
				Request: &models.StructuredRequest{
					Method: "GET",
					URL:    server.URL + "/not-found",
				},
				Assertions: []models.Assertion{
					{Type: models.AssertionStatus, Value: "404"},
				},
			},
		},
	}

	// Create runner and execute
	runner := NewRunner(5*time.Second, "")
	ctx := context.Background()
	result, err := runner.Run(ctx, suite)

	if err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	if result.TotalTests != 2 {
		t.Errorf("Expected 2 tests, got %d", result.TotalTests)
	}

	if result.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", result.PassedTests)
	}

	if result.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests, got %d", result.FailedTests)
	}
}

func TestRunner_Integration_Parallel(t *testing.T) {
	// Create test HTTP server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	// Create test suite with multiple tests
	tests := make([]models.Test, 5)
	for i := 0; i < 5; i++ {
		tests[i] = models.Test{
			Name: "Parallel test",
			Request: &models.StructuredRequest{
				Method: "GET",
				URL:    server.URL + "/test",
			},
			Assertions: []models.Assertion{
				{Type: models.AssertionStatus, Value: "200"},
			},
		}
	}

	suite := &models.TestSuite{Tests: tests}

	// Create runner and execute in parallel
	runner := NewRunner(5*time.Second, "")
	ctx := context.Background()
	result, err := runner.RunParallel(ctx, suite, 3, false)

	if err != nil {
		t.Fatalf("Runner.RunParallel failed: %v", err)
	}

	if result.TotalTests != 5 {
		t.Errorf("Expected 5 tests, got %d", result.TotalTests)
	}

	if result.PassedTests != 5 {
		t.Errorf("Expected 5 passed tests, got %d", result.PassedTests)
	}

	// Verify all tests were actually called
	if callCount != 5 {
		t.Errorf("Expected 5 HTTP calls, got %d", callCount)
	}
}

func TestRunner_Integration_FailFast(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create test suite with multiple failing tests
	tests := make([]models.Test, 5)
	for i := 0; i < 5; i++ {
		tests[i] = models.Test{
			Name: "Failing test",
			Request: &models.StructuredRequest{
				Method: "GET",
				URL:    server.URL + "/test",
			},
			Assertions: []models.Assertion{
				{Type: models.AssertionStatus, Value: "200"}, // Will fail
			},
		}
	}

	suite := &models.TestSuite{Tests: tests}

	// Create runner and execute with fail-fast
	runner := NewRunner(5*time.Second, "")
	ctx := context.Background()
	result, err := runner.RunParallel(ctx, suite, 2, true) // fail-fast enabled

	if err != nil {
		t.Fatalf("Runner.RunParallel failed: %v", err)
	}

	// With fail-fast, we should stop after first failure
	// But due to parallel execution, might execute a few tests
	if result.TotalTests == 5 {
		t.Error("Expected fail-fast to stop early, but all 5 tests executed")
	}

	if result.FailedTests == 0 {
		t.Error("Expected at least 1 failed test")
	}
}

func TestRunner_Integration_Redirect(t *testing.T) {
	// Create test HTTP server with redirect
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/end", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("redirected"))
	}))
	defer server.Close()

	// Test with redirect following disabled
	maxRedir := 0
	suite := &models.TestSuite{
		Tests: []models.Test{
			{
				Name: "No redirect test",
				Request: &models.StructuredRequest{
					Method: "GET",
					URL:    server.URL + "/start",
				},
				MaxRedirects: &maxRedir,
				Assertions: []models.Assertion{
					{Type: models.AssertionStatus, Value: "302"},
				},
			},
		},
	}

	runner := NewRunner(5*time.Second, "")
	ctx := context.Background()
	result, err := runner.Run(ctx, suite)

	if err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	if result.PassedTests != 1 {
		t.Errorf("Expected 1 passed test (catching redirect), got %d", result.PassedTests)
	}
}

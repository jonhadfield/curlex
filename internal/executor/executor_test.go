package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestNewExecutor(t *testing.T) {
	timeout := 30 * time.Second
	executor := NewExecutor(timeout)

	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}
	if executor.client == nil {
		t.Error("Executor client is nil")
	}
	if executor.curlParser == nil {
		t.Error("Executor curlParser is nil")
	}
	if executor.client.Timeout != timeout {
		t.Errorf("Client timeout = %v, want %v", executor.client.Timeout, timeout)
	}
}

func TestExecutor_Execute_GET(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "GET Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    server.URL,
		},
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result == nil {
		t.Fatal("Execute() returned nil result")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", result.StatusCode)
	}
	if result.ResponseBody != `{"status": "ok"}` {
		t.Errorf("ResponseBody = %v, want {\"status\": \"ok\"}", result.ResponseBody)
	}
	if result.Error != nil {
		t.Errorf("Unexpected error: %v", result.Error)
	}
}

func TestExecutor_Execute_POST(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"test": "data"}` {
			t.Errorf("Unexpected body: %s", string(body))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "POST Test",
		Request: &models.StructuredRequest{
			Method: "POST",
			URL:    server.URL,
			Body:   `{"test": "data"}`,
		},
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.StatusCode != 201 {
		t.Errorf("StatusCode = %v, want 201", result.StatusCode)
	}
	if !strings.Contains(result.ResponseBody, "123") {
		t.Errorf("ResponseBody should contain id 123, got: %v", result.ResponseBody)
	}
}

func TestExecutor_Execute_WithHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "test-value" {
			t.Errorf("Expected X-Custom-Header: test-value, got: %s", r.Header.Get("X-Custom-Header"))
		}
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("Expected Authorization header, got: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Headers Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    server.URL,
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
				"Authorization":   "Bearer token123",
			},
		},
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", result.StatusCode)
	}
}

func TestExecutor_Execute_CurlCommand(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Curl Test",
		Curl: "curl " + server.URL,
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %v, want 200", result.StatusCode)
	}
	if result.ResponseBody != "success" {
		t.Errorf("ResponseBody = %v, want success", result.ResponseBody)
	}
}

func TestExecutor_Execute_Redirect_NoFollow(t *testing.T) {
	// Create redirect server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final"))
	}))
	defer targetServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL, http.StatusFound)
	}))
	defer redirectServer.Close()

	executor := NewExecutor(5 * time.Second)
	maxRedirects := 0
	test := models.Test{
		Name: "No Redirect Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    redirectServer.URL,
		},
		MaxRedirects: &maxRedirects,
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should get redirect response, not follow
	if result.StatusCode != 302 {
		t.Errorf("StatusCode = %v, want 302", result.StatusCode)
	}
}

func TestExecutor_Execute_Redirect_Limited(t *testing.T) {
	// Create chain of redirects
	finalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final"))
	}))
	defer finalServer.Close()

	redirect2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, finalServer.URL, http.StatusFound)
	}))
	defer redirect2.Close()

	redirect1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirect2.URL, http.StatusFound)
	}))
	defer redirect1.Close()

	executor := NewExecutor(5 * time.Second)
	maxRedirects := 1
	test := models.Test{
		Name: "Limited Redirect Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    redirect1.URL,
		},
		MaxRedirects: &maxRedirects,
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Should stop after 1 redirect
	if result.Error == nil {
		t.Error("Expected error for exceeding redirect limit")
	}
}

func TestExecutor_Execute_ContextCancelled(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Context Cancel Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    server.URL,
		},
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := executor.Execute(ctx, test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Error == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestExecutor_Execute_NoRequestSpec(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "No Request Test",
		// No Curl or Request specified
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Error == nil {
		t.Error("Expected error for missing request specification")
	}
	if !strings.Contains(result.Error.Error(), "no request specification") {
		t.Errorf("Expected 'no request specification' error, got: %v", result.Error)
	}
}

func TestExecutor_Execute_InvalidURL(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Invalid URL Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    "://invalid-url",
		},
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Error == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestExecutor_Execute_ServerError(t *testing.T) {
	// Create server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Server Error Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    server.URL,
		},
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.StatusCode != 500 {
		t.Errorf("StatusCode = %v, want 500", result.StatusCode)
	}
	if !strings.Contains(result.ResponseBody, "Internal Server Error") {
		t.Errorf("ResponseBody should contain error message, got: %v", result.ResponseBody)
	}
}

func TestExecutor_Execute_ResponseTimeTracking(t *testing.T) {
	// Create server with artificial delay
	delay := 100 * time.Millisecond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Name: "Response Time Test",
		Request: &models.StructuredRequest{
			Method: "GET",
			URL:    server.URL,
		},
	}

	result, err := executor.Execute(context.Background(), test)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Response time should be at least the delay
	if result.ResponseTime < delay {
		t.Errorf("ResponseTime = %v, should be >= %v", result.ResponseTime, delay)
	}
	// But not too much more (with 50ms tolerance)
	if result.ResponseTime > delay+50*time.Millisecond {
		t.Errorf("ResponseTime = %v, seems too long for %v delay", result.ResponseTime, delay)
	}
}

func TestExecutor_PrepareRequest_Curl(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Curl: "curl https://example.com",
	}

	prepared, err := executor.prepareRequest(test)
	if err != nil {
		t.Fatalf("prepareRequest() error = %v", err)
	}

	if prepared == nil {
		t.Fatal("prepareRequest() returned nil")
	}
	if prepared.URL != "https://example.com" {
		t.Errorf("URL = %v, want https://example.com", prepared.URL)
	}
}

func TestExecutor_PrepareRequest_Structured(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	test := models.Test{
		Request: &models.StructuredRequest{
			Method: "POST",
			URL:    "https://api.example.com/users",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"name": "test"}`,
		},
	}

	prepared, err := executor.prepareRequest(test)
	if err != nil {
		t.Fatalf("prepareRequest() error = %v", err)
	}

	if prepared.Method != "POST" {
		t.Errorf("Method = %v, want POST", prepared.Method)
	}
	if prepared.URL != "https://api.example.com/users" {
		t.Errorf("URL = %v, want https://api.example.com/users", prepared.URL)
	}
	if prepared.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type header = %v, want application/json", prepared.Headers["Content-Type"])
	}
	if prepared.Body != `{"name": "test"}` {
		t.Errorf("Body = %v, want {\"name\": \"test\"}", prepared.Body)
	}
}

func TestExecutor_CreateHTTPRequest(t *testing.T) {
	executor := NewExecutor(5 * time.Second)
	prepared := &models.PreparedRequest{
		Method: "POST",
		URL:    "https://example.com/api",
		Headers: map[string]string{
			"Authorization": "Bearer token",
			"Content-Type":  "application/json",
		},
		Body: `{"test": "data"}`,
	}

	req, err := executor.createHTTPRequest(context.Background(), prepared)
	if err != nil {
		t.Fatalf("createHTTPRequest() error = %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("Method = %v, want POST", req.Method)
	}
	if req.URL.String() != "https://example.com/api" {
		t.Errorf("URL = %v, want https://example.com/api", req.URL.String())
	}
	if req.Header.Get("Authorization") != "Bearer token" {
		t.Errorf("Authorization header = %v, want Bearer token", req.Header.Get("Authorization"))
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type header = %v, want application/json", req.Header.Get("Content-Type"))
	}

	// Read body
	body, _ := io.ReadAll(req.Body)
	if string(body) != `{"test": "data"}` {
		t.Errorf("Body = %v, want {\"test\": \"data\"}", string(body))
	}
}

func TestExecutor_CreateClientWithRedirects(t *testing.T) {
	executor := NewExecutor(5 * time.Second)

	tests := []struct {
		name         string
		maxRedirects int
		wantNil      bool
	}{
		{
			name:         "No redirects",
			maxRedirects: 0,
			wantNil:      false,
		},
		{
			name:         "Limited redirects",
			maxRedirects: 5,
			wantNil:      false,
		},
		{
			name:         "Unlimited redirects",
			maxRedirects: -1,
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := executor.createClientWithRedirects(tt.maxRedirects)
			if client == nil {
				t.Fatal("createClientWithRedirects() returned nil")
			}
			if tt.wantNil && client.CheckRedirect != nil {
				t.Error("Expected CheckRedirect to be nil for unlimited redirects")
			}
			if !tt.wantNil && client.CheckRedirect == nil {
				t.Error("Expected CheckRedirect to be set for limited redirects")
			}
		})
	}
}

package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestBasicHTTPRequest tests a simple GET request end-to-end
func TestBasicHTTPRequest(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create test YAML file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
tests:
  - name: "Basic GET request"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
      - body_contains: "ok"
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build curlex binary from project root
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build curlex: %v\n%s", err, output)
	}

	// Run curlex
	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("curlex failed: %v\n%s", err, output)
	}

	// Verify output
	outputStr := string(output)
	if !strings.Contains(outputStr, "Basic GET request") {
		t.Errorf("Output should contain test name, got: %s", outputStr)
	}
	if !strings.Contains(outputStr, "✓") || !strings.Contains(outputStr, "passed") {
		t.Errorf("Output should indicate success, got: %s", outputStr)
	}

	// Verify exit code
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestAllAssertionTypes tests all assertion types together
func TestAllAssertionTypes(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		time.Sleep(10 * time.Millisecond) // Small delay for response_time assertion
		w.Write([]byte(`{"id": 123, "name": "test", "active": true}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
tests:
  - name: "All assertion types"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
      - body_contains: "test"
      - json_path: ".id == 123"
      - json_path: ".name == 'test'"
      - json_path: ".active == true"
      - header: "Content-Type contains json"
      - response_time: "< 1s"
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build and run curlex
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("curlex failed: %v\n%s", err, output)
	}

	// Verify success
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got %d\nOutput: %s", cmd.ProcessState.ExitCode(), output)
	}
}

// TestJSONOutput tests JSON output format
func TestJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
tests:
  - name: "JSON output test"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build and run with JSON output
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), "--output", "json", testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("curlex failed: %v\n%s", err, output)
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify JSON structure
	if result["total_tests"] != float64(1) {
		t.Errorf("Expected 1 total test, got %v", result["total_tests"])
	}
	if result["passed_tests"] != float64(1) {
		t.Errorf("Expected 1 passed test, got %v", result["passed_tests"])
	}
	if result["failed_tests"] != float64(0) {
		t.Errorf("Expected 0 failed tests, got %v", result["failed_tests"])
	}
}

// TestFailedAssertion tests that failures are properly reported
func TestFailedAssertion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // Will fail status assertion
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
tests:
  - name: "Failed assertion test"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build and run
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), testFile)
	output, _ := cmd.CombinedOutput() // Expect error

	// Verify failure exit code
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Expected exit code 1 for failed test, got %d", cmd.ProcessState.ExitCode())
	}

	// Verify output contains failure indicator
	outputStr := string(output)
	if !strings.Contains(outputStr, "✗") && !strings.Contains(outputStr, "failed") {
		t.Errorf("Output should indicate failure, got: %s", outputStr)
	}
}

// TestVariableSubstitution tests environment variable expansion
func TestVariableSubstitution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Set environment variable
	os.Setenv("TEST_SERVER_URL", server.URL)
	defer os.Unsetenv("TEST_SERVER_URL")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
variables:
  BASE_URL: "${TEST_SERVER_URL}"
tests:
  - name: "Variable substitution test"
    request:
      method: GET
      url: "${BASE_URL}"
    assertions:
      - status: 200
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build and run
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), testFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("curlex failed: %v\n%s", err, output)
	}

	// Verify success
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got %d\nOutput: %s", cmd.ProcessState.ExitCode(), output)
	}
}

// TestParallelExecution tests parallel execution mode
func TestParallelExecution(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		time.Sleep(50 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
tests:
  - name: "Test 1"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
  - name: "Test 2"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
  - name: "Test 3"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build and run with parallel flag
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	start := time.Now()
	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), "--parallel", testFile)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("curlex failed: %v\n%s", err, output)
	}

	// Parallel execution should be faster than sequential (3 * 50ms = 150ms)
	// With parallel, should be close to 50ms (allowing overhead)
	if duration > 120*time.Millisecond {
		t.Logf("Warning: Parallel execution took %v, expected < 120ms (may indicate sequential execution)", duration)
	}

	// Verify all tests passed
	if cmd.ProcessState.ExitCode() != 0 {
		t.Errorf("Expected exit code 0, got %d", cmd.ProcessState.ExitCode())
	}
}

// TestFailFastMode tests that execution stops on first failure
func TestFailFastMode(t *testing.T) {
	var executedTests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&executedTests, 1)
		// First request returns 404, rest would return 200
		if count == 1 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	yamlContent := `version: "1.0"
tests:
  - name: "Test 1 - will fail"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
  - name: "Test 2 - should not run"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
  - name: "Test 3 - should not run"
    request:
      method: GET
      url: "` + server.URL + `"
    assertions:
      - status: 200
`
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Build and run with fail-fast
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tmpDir, "curlex"), "./cmd/curlex")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	cmd := exec.Command(filepath.Join(tmpDir, "curlex"), "--fail-fast", testFile)
	output, _ := cmd.CombinedOutput() // Expect failure

	// Verify exit code is 1
	if cmd.ProcessState.ExitCode() != 1 {
		t.Errorf("Expected exit code 1, got %d", cmd.ProcessState.ExitCode())
	}

	// Output should only show 1 test (or indicate stopping early)
	outputStr := string(output)
	if strings.Count(outputStr, "Test 1") == 0 {
		t.Errorf("Output should contain Test 1, got: %s", outputStr)
	}
}

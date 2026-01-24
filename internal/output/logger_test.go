package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestRequestLogger_LogTest(t *testing.T) {
	tmpDir := t.TempDir()

	logger := NewRequestLogger(tmpDir)

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Request Logging",
		},
		StatusCode:   200,
		ResponseTime: 150 * time.Millisecond,
		ResponseBody: `{"status": "ok"}`,
		Headers:      map[string][]string{"Content-Type": {"application/json"}},
		Success:      true,
	}

	preparedReq := &models.PreparedRequest{
		Method:  "GET",
		URL:     "https://api.example.com/users",
		Headers: map[string]string{"Accept": "application/json"},
		Body:    "",
	}

	err := logger.LogTest(result, preparedReq)
	if err != nil {
		t.Fatalf("LogTest() error = %v", err)
	}

	// Check that log file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected 1 log file, got %d", len(files))
	}

	// Read log file
	logPath := filepath.Join(tmpDir, files[0].Name())
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	logContent := string(content)

	// Verify log contains key information
	if !strings.Contains(logContent, "=== REQUEST ===") {
		t.Error("Log should contain REQUEST section")
	}
	if !strings.Contains(logContent, "=== RESPONSE ===") {
		t.Error("Log should contain RESPONSE section")
	}
	if !strings.Contains(logContent, "GET https://api.example.com/users") {
		t.Error("Log should contain request method and URL")
	}
	if !strings.Contains(logContent, "Status: 200") {
		t.Error("Log should contain status code")
	}
	if !strings.Contains(logContent, "=== ASSERTIONS ===") {
		t.Error("Log should contain ASSERTIONS section")
	}
}

func TestRequestLogger_LogTest_WithFailures(t *testing.T) {
	tmpDir := t.TempDir()
	logger := NewRequestLogger(tmpDir)

	result := models.TestResult{
		Test: models.Test{
			Name: "Failed Test",
		},
		StatusCode:   404,
		ResponseTime: 200 * time.Millisecond,
		Success:      false,
		Failures: []models.AssertionFailure{
			{Type: "status", Expected: "200", Actual: "404"},
		},
	}

	err := logger.LogTest(result, nil)
	if err != nil {
		t.Fatalf("LogTest() error = %v", err)
	}

	// Read log file
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(tmpDir, files[0].Name())
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "assertion(s) failed") {
		t.Error("Log should indicate assertion failures")
	}
}

func TestRequestLogger_LogTest_NoLogging(t *testing.T) {
	// Logger with empty logDir should not create files
	logger := NewRequestLogger("")

	result := models.TestResult{
		Test:    models.Test{Name: "Test"},
		Success: true,
	}

	err := logger.LogTest(result, nil)
	if err != nil {
		t.Errorf("LogTest() with empty logDir should not error, got: %v", err)
	}
}

func TestRequestLogger_SensitiveHeaderMasking(t *testing.T) {
	tmpDir := t.TempDir()
	logger := NewRequestLogger(tmpDir)

	preparedReq := &models.PreparedRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Authorization": "Bearer secret_token_12345",
			"Cookie":        "session=xyz123",
			"X-API-Key":     "api_key_secret",
			"Content-Type":  "application/json",
		},
	}

	result := models.TestResult{
		Test:    models.Test{Name: "Sensitive Headers"},
		Success: true,
	}

	err := logger.LogTest(result, preparedReq)
	if err != nil {
		t.Fatal(err)
	}

	// Read log file
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(tmpDir, files[0].Name())
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	logContent := string(content)

	// Sensitive values should be redacted
	if strings.Contains(logContent, "secret_token_12345") {
		t.Error("Authorization header value should be redacted")
	}
	if strings.Contains(logContent, "session=xyz123") {
		t.Error("Cookie value should be redacted")
	}
	if strings.Contains(logContent, "api_key_secret") {
		t.Error("API key should be redacted")
	}

	// Should contain redaction marker
	if !strings.Contains(logContent, "***REDACTED***") {
		t.Error("Log should contain redaction marker for sensitive headers")
	}

	// Non-sensitive headers should not be redacted
	if !strings.Contains(logContent, "application/json") {
		t.Error("Non-sensitive headers should not be redacted")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple name",
			input: "test",
			want:  "test",
		},
		{
			name:  "Name with spaces",
			input: "my test name",
			want:  "my_test_name",
		},
		{
			name:  "Name with unsafe characters",
			input: "test/with:unsafe*chars?",
			want:  "test_with_unsafe_chars_",
		},
		{
			name:  "Very long name",
			input: "this is a very long test name that exceeds fifty characters limit",
			want:  "this_is_a_very_long_test_name_that_exceeds_fifty_c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.want)
			}
			if len(result) > 50 {
				t.Errorf("sanitizeFilename(%q) length %d exceeds 50", tt.input, len(result))
			}
		})
	}
}

func TestIsSensitiveHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   bool
	}{
		{"Authorization", "Authorization", true},
		{"authorization", "authorization", true},
		{"Cookie", "Cookie", true},
		{"cookie", "cookie", true},
		{"X-API-Key", "X-API-Key", true},
		{"x-api-key", "x-api-key", true},
		{"Token", "Token", true},
		{"Content-Type", "Content-Type", false},
		{"Accept", "Accept", false},
		{"User-Agent", "User-Agent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveHeader(tt.header)
			if result != tt.want {
				t.Errorf("isSensitiveHeader(%q) = %v, want %v", tt.header, result, tt.want)
			}
		})
	}
}

func TestFormatBody(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple body",
			input: "test",
			want:  "  test",
		},
		{
			name:  "Multi-line body",
			input: "line1\nline2",
			want:  "  line1\n  line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBody(tt.input)
			if result != tt.want {
				t.Errorf("formatBody(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

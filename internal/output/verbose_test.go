package output

import (
	"errors"
	"strings"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestNewVerboseFormatter(t *testing.T) {
	formatter := NewVerboseFormatter(false)
	if formatter == nil {
		t.Fatal("NewVerboseFormatter() returned nil")
	}
	if formatter.NoColor != false {
		t.Error("Expected NoColor to be false")
	}

	formatterNoColor := NewVerboseFormatter(true)
	if formatterNoColor.NoColor != true {
		t.Error("Expected NoColor to be true")
	}
}

func TestVerboseFormatter_FormatResult_Success(t *testing.T) {
	formatter := NewVerboseFormatter(true) // No color for easier testing

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Success",
		},
		PreparedRequest: &models.PreparedRequest{
			Method: "GET",
			URL:    "https://api.example.com/users",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			Body: "",
		},
		StatusCode:   200,
		ResponseTime: 150 * time.Millisecond,
		ResponseBody: `{"status": "ok"}`,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Success: true,
	}

	output := formatter.FormatResult(result)

	// Check for key sections
	if !strings.Contains(output, "✓ Test Success") {
		t.Error("Output should contain test name with checkmark")
	}
	if !strings.Contains(output, "REQUEST:") {
		t.Error("Output should contain REQUEST section")
	}
	if !strings.Contains(output, "GET https://api.example.com/users") {
		t.Error("Output should contain request method and URL")
	}
	if !strings.Contains(output, "RESPONSE:") {
		t.Error("Output should contain RESPONSE section")
	}
	if !strings.Contains(output, "Status: 200") {
		t.Error("Output should contain status code")
	}
	if !strings.Contains(output, "ASSERTIONS:") {
		t.Error("Output should contain ASSERTIONS section")
	}
	if !strings.Contains(output, "All assertions passed") {
		t.Error("Output should indicate all assertions passed")
	}
}

func TestVerboseFormatter_FormatResult_Failure(t *testing.T) {
	formatter := NewVerboseFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Failure",
		},
		StatusCode:   404,
		ResponseTime: 200 * time.Millisecond,
		ResponseBody: "Not Found",
		Success:      false,
		Failures: []models.AssertionFailure{
			{Type: "status", Expected: "200", Actual: "404"},
		},
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "✗ Test Failure") {
		t.Error("Output should contain test name with X mark")
	}
	if !strings.Contains(output, "Status: 404") {
		t.Error("Output should contain status code")
	}
	if !strings.Contains(output, "assertion(s) failed") {
		t.Error("Output should indicate assertion failures")
	}
}

func TestVerboseFormatter_FormatResult_WithError(t *testing.T) {
	formatter := NewVerboseFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Error",
		},
		StatusCode:   0,
		ResponseTime: 0,
		Success:      false,
		Error:        errors.New("connection timeout"),
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "ERROR:") {
		t.Error("Output should contain ERROR section")
	}
	if !strings.Contains(output, "connection timeout") {
		t.Error("Output should contain error message")
	}
}

func TestVerboseFormatter_FormatResult_SensitiveHeaders(t *testing.T) {
	formatter := NewVerboseFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "Sensitive Headers Test",
		},
		PreparedRequest: &models.PreparedRequest{
			Method: "GET",
			URL:    "https://api.example.com/users",
			Headers: map[string]string{
				"Authorization": "Bearer secret_token_12345",
				"Cookie":        "session=xyz123",
				"X-API-Key":     "api_key_secret",
				"Content-Type":  "application/json",
			},
		},
		StatusCode: 200,
		Success:    true,
	}

	output := formatter.FormatResult(result)

	// Sensitive values should be redacted
	if strings.Contains(output, "secret_token_12345") {
		t.Error("Authorization header value should be redacted")
	}
	if strings.Contains(output, "session=xyz123") {
		t.Error("Cookie value should be redacted")
	}
	if strings.Contains(output, "api_key_secret") {
		t.Error("API key should be redacted")
	}

	// Should contain redaction marker
	if !strings.Contains(output, "***REDACTED***") {
		t.Error("Output should contain redaction marker for sensitive headers")
	}

	// Non-sensitive headers should not be redacted
	if !strings.Contains(output, "application/json") {
		t.Error("Non-sensitive headers should not be redacted")
	}
}

func TestVerboseFormatter_FormatResult_LongBody(t *testing.T) {
	formatter := NewVerboseFormatter(true)

	longRequestBody := strings.Repeat("a", 250)
	longResponseBody := strings.Repeat("b", 350)

	result := models.TestResult{
		Test: models.Test{
			Name: "Long Body Test",
		},
		PreparedRequest: &models.PreparedRequest{
			Method: "POST",
			URL:    "https://api.example.com/data",
			Body:   longRequestBody,
		},
		StatusCode:   200,
		ResponseTime: 100 * time.Millisecond,
		ResponseBody: longResponseBody,
		Success:      true,
	}

	output := formatter.FormatResult(result)

	// Request body should be truncated at 200 chars
	if !strings.Contains(output, "...") {
		t.Error("Long request body should be truncated with ellipsis")
	}
	// Should show first 200 chars of request body
	if !strings.Contains(output, longRequestBody[:100]) {
		t.Error("Output should contain beginning of request body")
	}

	// Response body should be truncated at 300 chars
	if !strings.Contains(output, "first 300 chars") {
		t.Error("Output should indicate response body truncation")
	}
	if !strings.Contains(output, longResponseBody[:100]) {
		t.Error("Output should contain beginning of response body")
	}
}

func TestVerboseFormatter_FormatResult_WithResponseHeaders(t *testing.T) {
	formatter := NewVerboseFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "Response Headers Test",
		},
		StatusCode:   200,
		ResponseTime: 100 * time.Millisecond,
		Headers: map[string][]string{
			"Content-Type":  {"application/json"},
			"Cache-Control": {"no-cache", "no-store"},
		},
		Success: true,
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "Content-Type: application/json") {
		t.Error("Output should contain response headers")
	}
	if !strings.Contains(output, "Cache-Control: no-cache") {
		t.Error("Output should contain all header values")
	}
	if !strings.Contains(output, "Cache-Control: no-store") {
		t.Error("Output should contain all header values")
	}
}

func TestVerboseFormatter_FormatResult_NoRequestDetails(t *testing.T) {
	formatter := NewVerboseFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "No Request Details",
		},
		PreparedRequest: nil, // No request details
		StatusCode:      200,
		ResponseTime:    100 * time.Millisecond,
		Success:         true,
	}

	output := formatter.FormatResult(result)

	// Should still have RESPONSE section
	if !strings.Contains(output, "RESPONSE:") {
		t.Error("Output should contain RESPONSE section even without request details")
	}
	if !strings.Contains(output, "Status: 200") {
		t.Error("Output should contain status code")
	}
}

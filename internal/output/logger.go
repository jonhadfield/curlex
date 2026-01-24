package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"curlex/internal/models"
)

// RequestLogger logs full request/response details to files
type RequestLogger struct {
	logDir string
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(logDir string) *RequestLogger {
	return &RequestLogger{
		logDir: logDir,
	}
}

// LogTest saves request and response details to a log file
func (l *RequestLogger) LogTest(result models.TestResult, preparedReq *models.PreparedRequest) error {
	if l.logDir == "" {
		return nil // Logging not enabled
	}

	// Ensure log directory exists
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate filename: YYYY-MM-DD_HH-MM-SS_test-name.log
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	safeName := sanitizeFilename(result.Test.Name)
	filename := fmt.Sprintf("%s_%s.log", timestamp, safeName)
	filepath := filepath.Join(l.logDir, filename)

	// Build log content
	var content strings.Builder

	// === Request Section ===
	content.WriteString("=== REQUEST ===\n")
	if preparedReq != nil {
		content.WriteString(fmt.Sprintf("%s %s\n", preparedReq.Method, preparedReq.URL))
		if len(preparedReq.Headers) > 0 {
			content.WriteString("\nHeaders:\n")
			for key, value := range preparedReq.Headers {
				// Mask sensitive headers
				displayValue := value
				if isSensitiveHeader(key) {
					displayValue = "***REDACTED***"
				}
				content.WriteString(fmt.Sprintf("  %s: %s\n", key, displayValue))
			}
		}
		if preparedReq.Body != "" {
			content.WriteString("\nBody:\n")
			content.WriteString(formatBody(preparedReq.Body))
			content.WriteString("\n")
		}
	}

	// === Response Section ===
	content.WriteString("\n=== RESPONSE ===\n")
	content.WriteString(fmt.Sprintf("Status: %d (%dms)\n", result.StatusCode, result.ResponseTime.Milliseconds()))

	if len(result.Headers) > 0 {
		content.WriteString("\nHeaders:\n")
		for key, values := range result.Headers {
			for _, value := range values {
				content.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			}
		}
	}

	if result.ResponseBody != "" {
		content.WriteString("\nBody:\n")
		content.WriteString(formatBody(result.ResponseBody))
		content.WriteString("\n")
	}

	// === Assertions Section ===
	content.WriteString("\n=== ASSERTIONS ===\n")
	if len(result.Failures) == 0 {
		content.WriteString("✓ All assertions passed\n")
	} else {
		content.WriteString(fmt.Sprintf("✗ %d assertion(s) failed:\n", len(result.Failures)))
		for _, failure := range result.Failures {
			content.WriteString(fmt.Sprintf("  • %s\n", failure.String()))
		}
	}

	// === Error Section ===
	if result.Error != nil {
		content.WriteString("\n=== ERROR ===\n")
		content.WriteString(result.Error.Error())
		content.WriteString("\n")
	}

	// Write to file
	if err := os.WriteFile(filepath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}

// sanitizeFilename removes characters that are unsafe for filenames
func sanitizeFilename(name string) string {
	// Replace unsafe characters with underscores
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	safe := name
	for _, char := range unsafe {
		safe = strings.ReplaceAll(safe, char, "_")
	}
	// Limit length to 50 characters
	if len(safe) > 50 {
		safe = safe[:50]
	}
	return safe
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(key string) bool {
	lower := strings.ToLower(key)
	sensitive := []string{"authorization", "cookie", "api-key", "x-api-key", "token"}
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// formatBody attempts to format JSON bodies with indentation
func formatBody(body string) string {
	// For now, just return the body as-is
	// Future: could pretty-print JSON
	return "  " + strings.ReplaceAll(body, "\n", "\n  ")
}

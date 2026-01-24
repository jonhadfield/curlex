package output

import (
	"strings"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestQuietFormatter_AllPassed(t *testing.T) {
	formatter := NewQuietFormatter(true) // No color

	results := []models.TestResult{
		{Success: true},
		{Success: true},
		{Success: true},
	}

	output := formatter.FormatSummary(results, 1000*time.Millisecond)

	if !strings.Contains(output, "3/3 passed") {
		t.Errorf("Expected '3/3 passed' in output, got: %s", output)
	}

	if !strings.Contains(output, "1000ms") {
		t.Errorf("Expected '1000ms' in output, got: %s", output)
	}

	if !strings.Contains(output, "✓") {
		t.Errorf("Expected '✓' in output, got: %s", output)
	}
}

func TestQuietFormatter_SomeFailures(t *testing.T) {
	formatter := NewQuietFormatter(true) // No color

	results := []models.TestResult{
		{Success: true},
		{Success: false},
		{Success: true},
	}

	output := formatter.FormatSummary(results, 500*time.Millisecond)

	if !strings.Contains(output, "1/3 failed") {
		t.Errorf("Expected '1/3 failed' in output, got: %s", output)
	}

	if !strings.Contains(output, "2 passed") {
		t.Errorf("Expected '2 passed' in output, got: %s", output)
	}

	if !strings.Contains(output, "500ms") {
		t.Errorf("Expected '500ms' in output, got: %s", output)
	}

	if !strings.Contains(output, "✗") {
		t.Errorf("Expected '✗' in output, got: %s", output)
	}
}

func TestQuietFormatter_WithColor(t *testing.T) {
	formatter := NewQuietFormatter(false) // With color

	results := []models.TestResult{
		{Success: true},
	}

	output := formatter.FormatSummary(results, 100*time.Millisecond)

	// Should contain color codes
	if !strings.Contains(output, ColorGreen) {
		t.Error("Expected color codes in output")
	}
}

func TestQuietFormatter_NoColor(t *testing.T) {
	formatter := NewQuietFormatter(true) // No color

	results := []models.TestResult{
		{Success: true},
	}

	output := formatter.FormatSummary(results, 100*time.Millisecond)

	// Should NOT contain color codes
	if strings.Contains(output, "\033[") {
		t.Error("Expected no color codes in output")
	}
}

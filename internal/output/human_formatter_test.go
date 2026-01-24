package output

import (
	"errors"
	"strings"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestNewHumanFormatter(t *testing.T) {
	formatter := NewHumanFormatter(false)
	if formatter == nil {
		t.Fatal("NewHumanFormatter() returned nil")
	}
	if formatter.NoColor != false {
		t.Error("Expected NoColor to be false")
	}

	formatterNoColor := NewHumanFormatter(true)
	if formatterNoColor.NoColor != true {
		t.Error("Expected NoColor to be true")
	}
}

func TestHumanFormatter_FormatResult_Success(t *testing.T) {
	formatter := NewHumanFormatter(true) // No color for easier testing

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Success",
		},
		StatusCode:   200,
		ResponseTime: 150 * time.Millisecond,
		Success:      true,
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "✓") {
		t.Error("Output should contain checkmark for success")
	}
	if !strings.Contains(output, "Test Success") {
		t.Error("Output should contain test name")
	}
	if !strings.Contains(output, "200") {
		t.Error("Output should contain status code")
	}
	if !strings.Contains(output, "150ms") {
		t.Error("Output should contain response time")
	}
}

func TestHumanFormatter_FormatResult_Failure(t *testing.T) {
	formatter := NewHumanFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Failure",
		},
		StatusCode:   404,
		ResponseTime: 200 * time.Millisecond,
		Success:      false,
		Failures: []models.AssertionFailure{
			{Expected: "200", Actual: "404"},
		},
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "✗") {
		t.Error("Output should contain X mark for failure")
	}
	if !strings.Contains(output, "Test Failure") {
		t.Error("Output should contain test name")
	}
	if !strings.Contains(output, "404") {
		t.Error("Output should contain status code")
	}
}

func TestHumanFormatter_FormatResult_WithError(t *testing.T) {
	formatter := NewHumanFormatter(true)

	result := models.TestResult{
		Test: models.Test{
			Name: "Test Error",
		},
		Success: false,
		Error:   errors.New("connection timeout"),
	}

	output := formatter.FormatResult(result)

	if !strings.Contains(output, "✗") {
		t.Error("Output should contain X mark for error")
	}
	if !strings.Contains(output, "connection timeout") {
		t.Error("Output should contain error message")
	}
}

func TestHumanFormatter_FormatSummary(t *testing.T) {
	formatter := NewHumanFormatter(true)

	results := []models.TestResult{
		{Test: models.Test{Name: "Test 1"}, Success: true},
		{Test: models.Test{Name: "Test 2"}, Success: true},
		{Test: models.Test{Name: "Test 3"}, Success: true},
		{Test: models.Test{Name: "Test 4"}, Success: true},
		{Test: models.Test{Name: "Test 5"}, Success: true},
		{Test: models.Test{Name: "Test 6"}, Success: true},
		{Test: models.Test{Name: "Test 7"}, Success: true},
		{Test: models.Test{Name: "Test 8"}, Success: true},
		{Test: models.Test{Name: "Test 9"}, Success: false},
		{Test: models.Test{Name: "Test 10"}, Success: false},
	}

	output := formatter.FormatSummary(results, 5*time.Second)

	if !strings.Contains(output, "10") {
		t.Error("Summary should contain total tests")
	}
	if !strings.Contains(output, "8") {
		t.Error("Summary should contain passed tests")
	}
	if !strings.Contains(output, "2") {
		t.Error("Summary should contain failed tests")
	}
	if !strings.Contains(output, "5") {
		t.Error("Summary should contain total time")
	}
}

func TestHumanFormatter_FormatSummary_AllPassed(t *testing.T) {
	formatter := NewHumanFormatter(true)

	results := []models.TestResult{
		{Test: models.Test{Name: "Test 1"}, Success: true},
		{Test: models.Test{Name: "Test 2"}, Success: true},
		{Test: models.Test{Name: "Test 3"}, Success: true},
		{Test: models.Test{Name: "Test 4"}, Success: true},
		{Test: models.Test{Name: "Test 5"}, Success: true},
	}

	output := formatter.FormatSummary(results, 2*time.Second)

	if !strings.Contains(output, "5") {
		t.Error("Summary should contain test count")
	}
}

func TestHumanFormatter_Colorize(t *testing.T) {
	tests := []struct {
		name     string
		noColor  bool
		color    string
		text     string
		contains []string
	}{
		{
			name:     "with color",
			noColor:  false,
			color:    ColorRed,
			text:     "Error",
			contains: []string{"\033[31m", "Error", "\033[0m"},
		},
		{
			name:     "no color mode",
			noColor:  true,
			color:    ColorRed,
			text:     "Error",
			contains: []string{"Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewHumanFormatter(tt.noColor)
			output := formatter.colorize(tt.color, tt.text)

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("colorize() should contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

func TestHumanFormatter_Indent(t *testing.T) {
	formatter := NewHumanFormatter(true)

	tests := []struct {
		name   string
		text   string
		spaces int
		want   string
	}{
		{
			name:   "single line 2 spaces",
			text:   "test",
			spaces: 2,
			want:   "  test",
		},
		{
			name:   "single line 4 spaces",
			text:   "hello",
			spaces: 4,
			want:   "    hello",
		},
		{
			name:   "text with newline",
			text:   "line1\nline2",
			spaces: 2,
			want:   "  line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.indent(tt.text, tt.spaces)
			if result != tt.want {
				t.Errorf("indent() = %q, want %q", result, tt.want)
			}
		})
	}
}

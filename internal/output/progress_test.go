package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewProgress(t *testing.T) {
	progress := NewProgress(10, false)
	if progress == nil {
		t.Fatal("NewProgress() returned nil")
	}
	if progress.total != 10 {
		t.Errorf("total = %d, want 10", progress.total)
	}
	if progress.current != 0 {
		t.Errorf("current = %d, want 0", progress.current)
	}
	if progress.noColor != false {
		t.Error("Expected noColor to be false")
	}
	if progress.done == nil {
		t.Error("done channel should be initialized")
	}

	progressNoColor := NewProgress(5, true)
	if progressNoColor.noColor != true {
		t.Error("Expected noColor to be true")
	}
}

func TestProgress_Increment(t *testing.T) {
	progress := NewProgress(10, false)

	if progress.current != 0 {
		t.Errorf("initial current = %d, want 0", progress.current)
	}

	progress.Increment()
	if progress.current != 1 {
		t.Errorf("after Increment, current = %d, want 1", progress.current)
	}

	progress.Increment()
	progress.Increment()
	if progress.current != 3 {
		t.Errorf("after 3 Increments, current = %d, want 3", progress.current)
	}
}

func TestProgress_FormatProgressBar(t *testing.T) {
	tests := []struct {
		name     string
		current  int
		total    int
		noColor  bool
		spinner  string
		contains []string
	}{
		{
			name:     "No color mode - partial progress",
			current:  5,
			total:    10,
			noColor:  true,
			spinner:  "⠋",
			contains: []string{"⠋", "Running tests", "5/10", "50%"},
		},
		{
			name:     "No color mode - complete",
			current:  10,
			total:    10,
			noColor:  true,
			spinner:  "⠙",
			contains: []string{"⠙", "Running tests", "10/10", "100%"},
		},
		{
			name:     "No color mode - zero progress",
			current:  0,
			total:    20,
			noColor:  true,
			spinner:  "⠹",
			contains: []string{"⠹", "Running tests", "0/20", "0%"},
		},
		{
			name:     "Color mode - partial progress",
			current:  3,
			total:    6,
			noColor:  false,
			spinner:  "⠸",
			contains: []string{"⠸", "Running tests", "3/6", "50%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := NewProgress(tt.total, tt.noColor)
			output := progress.formatProgressBar(tt.current, tt.total, tt.spinner)

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("formatProgressBar() output should contain %q, got: %s", expected, output)
				}
			}

			// Check for progress bar characters
			if !strings.Contains(output, "█") && !strings.Contains(output, "░") {
				t.Error("formatProgressBar() should contain progress bar characters")
			}
		})
	}
}

func TestProgress_FormatSpinner(t *testing.T) {
	tests := []struct {
		name        string
		current     int
		noColor     bool
		spinner     string
		shouldMatch []string
	}{
		{
			name:        "No color mode - zero tests",
			current:     0,
			noColor:     true,
			spinner:     "⠋",
			shouldMatch: []string{"⠋", "Running tests", "0", "completed"},
		},
		{
			name:        "No color mode - some tests",
			current:     5,
			noColor:     true,
			spinner:     "⠙",
			shouldMatch: []string{"⠙", "Running tests", "5", "completed"},
		},
		{
			name:        "Color mode - some tests",
			current:     10,
			noColor:     false,
			spinner:     "⠹",
			shouldMatch: []string{"⠹", "Running tests", "10", "completed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := NewProgress(0, tt.noColor)
			output := progress.formatSpinner(tt.current, tt.spinner)

			for _, expected := range tt.shouldMatch {
				if !strings.Contains(output, expected) {
					t.Errorf("formatSpinner() output should contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

func TestProgress_StopWhenNotActive(t *testing.T) {
	progress := NewProgress(10, false)
	// Stop should not panic or hang when not active
	progress.Stop()
}

func TestProgress_Clear(t *testing.T) {
	var buf bytes.Buffer
	progress := NewProgress(10, false)
	progress.writer = &buf
	progress.isTerminal = true

	progress.clear()

	output := buf.String()
	// Should contain ANSI escape codes for clearing line
	if !strings.Contains(output, "\r") || !strings.Contains(output, "\033[K") {
		t.Error("clear() should output ANSI escape codes for clearing line")
	}
}

func TestProgress_ClearNonTerminal(t *testing.T) {
	var buf bytes.Buffer
	progress := NewProgress(10, false)
	progress.writer = &buf
	progress.isTerminal = false

	progress.clear()

	output := buf.String()
	// Should not output anything in non-terminal mode
	if output != "" {
		t.Error("clear() should not output anything in non-terminal mode")
	}
}

package output

import (
	"fmt"
	"time"

	"curlex/internal/models"
)

// QuietFormatter provides minimal output (summary only)
type QuietFormatter struct {
	NoColor bool
}

// NewQuietFormatter creates a new quiet formatter
func NewQuietFormatter(noColor bool) *QuietFormatter {
	return &QuietFormatter{
		NoColor: noColor,
	}
}

// FormatSummary outputs only the final summary
func (f *QuietFormatter) FormatSummary(results []models.TestResult, duration time.Duration) string {
	passed := 0
	failed := 0
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
		}
	}

	total := len(results)

	// Simple one-line output
	if failed == 0 {
		return f.colorize(ColorGreen, fmt.Sprintf("✓ %d/%d passed (%dms)\n", passed, total, duration.Milliseconds()))
	}
	return f.colorize(ColorRed, fmt.Sprintf("✗ %d/%d failed, %d passed (%dms)\n", failed, total, passed, duration.Milliseconds()))
}

// colorize applies color codes if colors are enabled
func (f *QuietFormatter) colorize(color, text string) string {
	if f.NoColor {
		return text
	}
	return color + text + ColorReset
}

package output

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"curlex/internal/models"
)

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

// HumanFormatter formats output for human-readable terminal display
type HumanFormatter struct {
	NoColor bool
}

// NewHumanFormatter creates a new human-readable formatter
func NewHumanFormatter(noColor bool) *HumanFormatter {
	return &HumanFormatter{
		NoColor: noColor,
	}
}

// FormatResult outputs a single test result
func (f *HumanFormatter) FormatResult(result models.TestResult) string {
	var sb strings.Builder

	// Test name with status icon
	if result.Success {
		sb.WriteString(f.colorize(ColorGreen, "✓"))
	} else {
		sb.WriteString(f.colorize(ColorRed, "✗"))
	}
	sb.WriteString(" ")
	sb.WriteString(f.colorize(ColorBold, result.Test.Name))
	sb.WriteString("\n")

	// Show error if present
	if result.Error != nil {
		sb.WriteString(f.indent(f.colorize(ColorRed, "Error: "+result.Error.Error()), 2))
		sb.WriteString("\n")
		return sb.String()
	}

	// Show status code and response time
	statusColor := ColorGreen
	if result.StatusCode >= 400 {
		statusColor = ColorRed
	} else if result.StatusCode >= 300 {
		statusColor = ColorYellow
	}

	sb.WriteString(f.indent(
		fmt.Sprintf("%s %s%s  %s%dms%s",
			f.colorize(ColorGray, "Status:"),
			f.colorize(statusColor, strconv.Itoa(result.StatusCode)),
			ColorReset,
			f.colorize(ColorGray, ""),
			result.ResponseTime.Milliseconds(),
			ColorReset,
		),
		2,
	))
	sb.WriteString("\n")

	// Show debug information if enabled
	if result.Test.Debug {
		// Show response headers
		sb.WriteString(f.indent(f.colorize(ColorBlue, "Headers:"), 2))
		sb.WriteString("\n")
		for key, values := range result.Headers {
			for _, value := range values {
				sb.WriteString(f.indent(fmt.Sprintf("%s: %s", key, value), 4))
				sb.WriteString("\n")
			}
		}

		// Show first 500 characters of response body
		sb.WriteString(f.indent(f.colorize(ColorBlue, "Body (first 500 chars):"), 2))
		sb.WriteString("\n")
		body := result.ResponseBody
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		// Indent each line of the body
		for _, line := range strings.Split(body, "\n") {
			sb.WriteString(f.indent(line, 4))
			sb.WriteString("\n")
		}
	}

	// Show assertion failures
	if len(result.Failures) > 0 {
		sb.WriteString(f.indent(f.colorize(ColorRed, "Failures:"), 2))
		sb.WriteString("\n")
		for _, failure := range result.Failures {
			sb.WriteString(f.indent(f.colorize(ColorRed, "• "+failure.String()), 4))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatSummary outputs the final summary
func (f *HumanFormatter) FormatSummary(results []models.TestResult, duration time.Duration) string {
	var sb strings.Builder

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

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", 50))
	sb.WriteString("\n")

	// Summary line
	if failed == 0 {
		sb.WriteString(f.colorize(ColorGreen+ColorBold, fmt.Sprintf("✓ All %d tests passed", total)))
	} else {
		sb.WriteString(f.colorize(ColorRed+ColorBold, fmt.Sprintf("✗ %d of %d tests failed", failed, total)))
	}
	sb.WriteString("\n")

	// Stats
	sb.WriteString(fmt.Sprintf("%sPassed:%s %d  ", f.colorize(ColorGray, ""), ColorReset, passed))
	if failed > 0 {
		sb.WriteString(fmt.Sprintf("%sFailed:%s %s%d%s  ", f.colorize(ColorGray, ""), ColorReset, ColorRed, failed, ColorReset))
	} else {
		sb.WriteString(fmt.Sprintf("%sFailed:%s %d  ", f.colorize(ColorGray, ""), ColorReset, failed))
	}
	sb.WriteString(fmt.Sprintf("%sTotal:%s %d  ", f.colorize(ColorGray, ""), ColorReset, total))
	sb.WriteString(fmt.Sprintf("%sTime:%s %dms\n", f.colorize(ColorGray, ""), ColorReset, duration.Milliseconds()))

	sb.WriteString(strings.Repeat("─", 50))
	sb.WriteString("\n")

	return sb.String()
}

// colorize applies color codes to text if colors are enabled
func (f *HumanFormatter) colorize(color, text string) string {
	if f.NoColor {
		return text
	}
	return color + text + ColorReset
}

// indent adds spaces to the beginning of text
func (f *HumanFormatter) indent(text string, spaces int) string {
	return strings.Repeat(" ", spaces) + text
}

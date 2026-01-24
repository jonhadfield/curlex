package output

import (
	"fmt"
	"strings"

	"curlex/internal/models"
)

// VerboseFormatter provides detailed output for troubleshooting
type VerboseFormatter struct {
	HumanFormatter
}

// NewVerboseFormatter creates a new verbose formatter
func NewVerboseFormatter(noColor bool) *VerboseFormatter {
	return &VerboseFormatter{
		HumanFormatter: HumanFormatter{NoColor: noColor},
	}
}

// FormatResult outputs detailed test result with full request/response info
func (f *VerboseFormatter) FormatResult(result models.TestResult) string {
	var sb strings.Builder

	// Test name with separator
	sb.WriteString(strings.Repeat("=", 60))
	sb.WriteString("\n")
	if result.Success {
		sb.WriteString(f.colorize(ColorGreen+ColorBold, "✓ "+result.Test.Name))
	} else {
		sb.WriteString(f.colorize(ColorRed+ColorBold, "✗ "+result.Test.Name))
	}
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", 60))
	sb.WriteString("\n\n")

	// Request details
	if result.PreparedRequest != nil {
		sb.WriteString(f.colorize(ColorBlue+ColorBold, "REQUEST:"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("  %s %s\n", result.PreparedRequest.Method, result.PreparedRequest.URL))

		if len(result.PreparedRequest.Headers) > 0 {
			sb.WriteString(f.colorize(ColorBlue, "  Headers:"))
			sb.WriteString("\n")
			for key, value := range result.PreparedRequest.Headers {
				// Mask sensitive headers
				displayValue := value
				if isSensitiveHeader(key) {
					displayValue = "***REDACTED***"
				}
				sb.WriteString(fmt.Sprintf("    %s: %s\n", key, displayValue))
			}
		}

		if result.PreparedRequest.Body != "" {
			sb.WriteString(f.colorize(ColorBlue, "  Body:"))
			sb.WriteString("\n")
			// Show first 200 characters
			body := result.PreparedRequest.Body
			if len(body) > 200 {
				body = body[:200] + "..."
			}
			sb.WriteString("    " + strings.ReplaceAll(body, "\n", "\n    ") + "\n")
		}
		sb.WriteString("\n")
	}

	// Response details
	sb.WriteString(f.colorize(ColorBlue+ColorBold, "RESPONSE:"))
	sb.WriteString("\n")

	// Status and timing
	statusColor := ColorGreen
	if result.StatusCode >= 400 {
		statusColor = ColorRed
	} else if result.StatusCode >= 300 {
		statusColor = ColorYellow
	}
	sb.WriteString(fmt.Sprintf("  Status: %s (%dms)\n",
		f.colorize(statusColor, fmt.Sprintf("%d", result.StatusCode)),
		result.ResponseTime.Milliseconds()))

	// Headers
	if len(result.Headers) > 0 {
		sb.WriteString(f.colorize(ColorBlue, "  Headers:"))
		sb.WriteString("\n")
		for key, values := range result.Headers {
			for _, value := range values {
				sb.WriteString(fmt.Sprintf("    %s: %s\n", key, value))
			}
		}
	}

	// Body
	if result.ResponseBody != "" {
		sb.WriteString(f.colorize(ColorBlue, "  Body (first 300 chars):"))
		sb.WriteString("\n")
		body := result.ResponseBody
		if len(body) > 300 {
			body = body[:300] + "..."
		}
		sb.WriteString("    " + strings.ReplaceAll(body, "\n", "\n    ") + "\n")
	}
	sb.WriteString("\n")

	// Assertions
	sb.WriteString(f.colorize(ColorBlue+ColorBold, "ASSERTIONS:"))
	sb.WriteString("\n")
	if len(result.Failures) == 0 {
		sb.WriteString(f.colorize(ColorGreen, "  ✓ All assertions passed"))
		sb.WriteString("\n")
	} else {
		sb.WriteString(f.colorize(ColorRed, fmt.Sprintf("  ✗ %d assertion(s) failed:", len(result.Failures))))
		sb.WriteString("\n")
		for _, failure := range result.Failures {
			sb.WriteString(f.colorize(ColorRed, "    • "+failure.String()))
			sb.WriteString("\n")
		}
	}

	// Error if present
	if result.Error != nil {
		sb.WriteString("\n")
		sb.WriteString(f.colorize(ColorRed+ColorBold, "ERROR:"))
		sb.WriteString("\n")
		sb.WriteString(f.colorize(ColorRed, "  "+result.Error.Error()))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

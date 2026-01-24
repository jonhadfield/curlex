package output

import (
	"encoding/json"
	"time"

	"curlex/internal/models"
)

// JSONFormatter formats test results as JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// JSONOutput represents the JSON output structure
type JSONOutput struct {
	Version     string           `json:"version"`
	TotalTests  int              `json:"total_tests"`
	PassedTests int              `json:"passed_tests"`
	FailedTests int              `json:"failed_tests"`
	TotalTime   string           `json:"total_time"`
	StartTime   string           `json:"start_time"`
	EndTime     string           `json:"end_time"`
	Tests       []JSONTestResult `json:"tests"`
}

// JSONTestResult represents a single test result in JSON format
type JSONTestResult struct {
	Name         string              `json:"name"`
	Success      bool                `json:"success"`
	StatusCode   int                 `json:"status_code,omitempty"`
	ResponseTime string              `json:"response_time,omitempty"`
	Error        string              `json:"error,omitempty"`
	Failures     []JSONFailure       `json:"failures,omitempty"`
	Request      *JSONRequest        `json:"request,omitempty"`
	Response     *JSONResponse       `json:"response,omitempty"`
}

// JSONRequest represents request details in JSON format
type JSONRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// JSONResponse represents response details in JSON format
type JSONResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
}

// JSONFailure represents an assertion failure in JSON format
type JSONFailure struct {
	Type     string `json:"type"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Message  string `json:"message"`
}

// Format converts suite results to JSON
func (f *JSONFormatter) Format(suiteResult *models.SuiteResult) string {
	output := JSONOutput{
		Version:     "1.0.0",
		TotalTests:  suiteResult.TotalTests,
		PassedTests: suiteResult.PassedTests,
		FailedTests: suiteResult.FailedTests,
		TotalTime:   formatDuration(suiteResult.TotalTime),
		StartTime:   suiteResult.StartTime.Format(time.RFC3339),
		EndTime:     suiteResult.EndTime.Format(time.RFC3339),
		Tests:       make([]JSONTestResult, 0, len(suiteResult.Results)),
	}

	for _, result := range suiteResult.Results {
		testResult := JSONTestResult{
			Name:         result.Test.Name,
			Success:      result.Success,
			StatusCode:   result.StatusCode,
			ResponseTime: formatDuration(result.ResponseTime),
		}

		if result.Error != nil {
			testResult.Error = result.Error.Error()
		}

		// Add failures
		if len(result.Failures) > 0 {
			testResult.Failures = make([]JSONFailure, 0, len(result.Failures))
			for _, failure := range result.Failures {
				testResult.Failures = append(testResult.Failures, JSONFailure{
					Type:     string(failure.Type),
					Expected: failure.Expected,
					Actual:   failure.Actual,
					Message:  failure.Message,
				})
			}
		}

		// Add request details
		if result.PreparedRequest != nil {
			testResult.Request = &JSONRequest{
				Method:  result.PreparedRequest.Method,
				URL:     result.PreparedRequest.URL,
				Headers: result.PreparedRequest.Headers,
				Body:    result.PreparedRequest.Body,
			}
		}

		// Add response details
		if result.StatusCode > 0 {
			testResult.Response = &JSONResponse{
				StatusCode: result.StatusCode,
				Headers:    result.Headers,
				Body:       result.ResponseBody,
			}
		}

		output.Tests = append(output.Tests, testResult)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal JSON"}`
	}

	return string(data) + "\n"
}

// formatDuration converts a duration to a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(time.Millisecond).String()
}

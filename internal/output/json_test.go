package output

import (
	"encoding/json"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestJSONFormatter_Format(t *testing.T) {
	formatter := NewJSONFormatter()

	suiteResult := &models.SuiteResult{
		TotalTests:  2,
		PassedTests: 1,
		FailedTests: 1,
		TotalTime:   1500 * time.Millisecond,
		StartTime:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2024, 1, 1, 12, 0, 1, 500000000, time.UTC),
		Results: []models.TestResult{
			{
				Test: models.Test{
					Name: "Success test",
				},
				Success:      true,
				StatusCode:   200,
				ResponseTime: 500 * time.Millisecond,
				PreparedRequest: &models.PreparedRequest{
					Method: "GET",
					URL:    "https://example.com",
				},
			},
			{
				Test: models.Test{
					Name: "Failed test",
				},
				Success:      false,
				StatusCode:   404,
				ResponseTime: 300 * time.Millisecond,
				Failures: []models.AssertionFailure{
					{
						Type:     models.AssertionStatus,
						Expected: "200",
						Actual:   "404",
						Message:  "expected status 200, got 404",
					},
				},
			},
		},
	}

	output := formatter.Format(suiteResult)

	// Parse JSON to validate structure
	var result JSONOutput
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Validate basic fields
	if result.TotalTests != 2 {
		t.Errorf("Expected total_tests 2, got %d", result.TotalTests)
	}

	if result.PassedTests != 1 {
		t.Errorf("Expected passed_tests 1, got %d", result.PassedTests)
	}

	if result.FailedTests != 1 {
		t.Errorf("Expected failed_tests 1, got %d", result.FailedTests)
	}

	// Validate test results
	if len(result.Tests) != 2 {
		t.Fatalf("Expected 2 test results, got %d", len(result.Tests))
	}

	// Check first test (success)
	if !result.Tests[0].Success {
		t.Error("Expected first test to be successful")
	}

	if result.Tests[0].StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", result.Tests[0].StatusCode)
	}

	// Check second test (failure)
	if result.Tests[1].Success {
		t.Error("Expected second test to fail")
	}

	if len(result.Tests[1].Failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(result.Tests[1].Failures))
	}
}

func TestJSONFormatter_ValidJSON(t *testing.T) {
	formatter := NewJSONFormatter()

	suiteResult := &models.SuiteResult{
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
		TotalTime:   0,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Results:     []models.TestResult{},
	}

	output := formatter.Format(suiteResult)

	// Verify it's valid JSON
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}
}

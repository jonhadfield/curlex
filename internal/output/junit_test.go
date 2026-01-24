package output

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"curlex/internal/models"
)

func TestJUnitFormatter_Format(t *testing.T) {
	formatter := NewJUnitFormatter()

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

	// Verify XML header
	if !strings.HasPrefix(output, xml.Header) {
		t.Error("Output should start with XML header")
	}

	// Parse XML to validate structure
	var result JUnitTestSuites
	err := xml.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JUnit XML: %v", err)
	}

	// Validate test suite
	if len(result.Suites) != 1 {
		t.Fatalf("Expected 1 test suite, got %d", len(result.Suites))
	}

	suite := result.Suites[0]
	if suite.Tests != 2 {
		t.Errorf("Expected 2 tests, got %d", suite.Tests)
	}

	if suite.Failures != 1 {
		t.Errorf("Expected 1 failure, got %d", suite.Failures)
	}

	// Validate test cases
	if len(suite.Cases) != 2 {
		t.Fatalf("Expected 2 test cases, got %d", len(suite.Cases))
	}

	// Check first test (success)
	if suite.Cases[0].Failure != nil {
		t.Error("Expected first test to have no failure")
	}

	// Check second test (failure)
	if suite.Cases[1].Failure == nil {
		t.Error("Expected second test to have failure")
	}
}

func TestJUnitFormatter_ValidXML(t *testing.T) {
	formatter := NewJUnitFormatter()

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

	// Verify it's valid XML
	var result JUnitTestSuites
	err := xml.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Output is not valid XML: %v", err)
	}
}

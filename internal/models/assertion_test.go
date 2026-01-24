package models

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAssertion_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedType  AssertionType
		expectedValue string
		shouldError   bool
	}{
		{
			name:          "status assertion",
			yaml:          "status: 200",
			expectedType:  AssertionStatus,
			expectedValue: "200",
			shouldError:   false,
		},
		{
			name:          "body assertion",
			yaml:          "body: '{\"test\": \"data\"}'",
			expectedType:  AssertionBody,
			expectedValue: `{"test": "data"}`,
			shouldError:   false,
		},
		{
			name:          "body_contains assertion",
			yaml:          "body_contains: success",
			expectedType:  AssertionBodyContains,
			expectedValue: "success",
			shouldError:   false,
		},
		{
			name:          "json_path assertion",
			yaml:          "json_path: '.data.id == 123'",
			expectedType:  AssertionJSONPath,
			expectedValue: ".data.id == 123",
			shouldError:   false,
		},
		{
			name:          "header assertion",
			yaml:          "header: 'Content-Type contains json'",
			expectedType:  AssertionHeader,
			expectedValue: "Content-Type contains json",
			shouldError:   false,
		},
		{
			name:          "response_time assertion",
			yaml:          "response_time: '< 500ms'",
			expectedType:  AssertionResponseTime,
			expectedValue: "< 500ms",
			shouldError:   false,
		},
		{
			name:        "unknown assertion type",
			yaml:        "unknown_type: value",
			shouldError: true,
		},
		{
			name:        "multiple keys",
			yaml:        "status: 200\nheader: 'test'",
			shouldError: true,
		},
		{
			name:        "invalid yaml",
			yaml:        "- this is a list",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var assertion Assertion
			err := yaml.Unmarshal([]byte(tt.yaml), &assertion)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if assertion.Type != tt.expectedType {
				t.Errorf("Type = %s, want %s", assertion.Type, tt.expectedType)
			}

			if assertion.Value != tt.expectedValue {
				t.Errorf("Value = %s, want %s", assertion.Value, tt.expectedValue)
			}
		})
	}
}

func TestAssertion_String(t *testing.T) {
	tests := []struct {
		name     string
		assertion Assertion
		expected string
	}{
		{
			name: "status assertion",
			assertion: Assertion{
				Type:  AssertionStatus,
				Value: "200",
			},
			expected: "status: 200",
		},
		{
			name: "json_path assertion",
			assertion: Assertion{
				Type:  AssertionJSONPath,
				Value: ".data.id == 123",
			},
			expected: "json_path: .data.id == 123",
		},
		{
			name: "header assertion",
			assertion: Assertion{
				Type:  AssertionHeader,
				Value: "Content-Type contains json",
			},
			expected: "header: Content-Type contains json",
		},
		{
			name: "response_time assertion",
			assertion: Assertion{
				Type:  AssertionResponseTime,
				Value: "< 500ms",
			},
			expected: "response_time: < 500ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.assertion.String()
			if result != tt.expected {
				t.Errorf("String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestAssertionFailure_String(t *testing.T) {
	tests := []struct {
		name     string
		failure  AssertionFailure
		contains []string
	}{
		{
			name: "with expected and actual",
			failure: AssertionFailure{
				Type:     "status",
				Expected: "200",
				Actual:   "404",
			},
			contains: []string{"expected", "200", "got", "404"},
		},
		{
			name: "with custom message",
			failure: AssertionFailure{
				Type:     "status",
				Expected: "200",
				Actual:   "404",
				Message:  "Status code mismatch",
			},
			contains: []string{"Status code mismatch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.failure.String()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("String() should contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

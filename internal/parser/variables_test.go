package parser

import (
	"os"
	"testing"

	"curlex/internal/models"
)

func TestVariableExpander_ExpandVariables(t *testing.T) {
	// Set environment variable for testing
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	tests := []struct {
		name        string
		suite       *models.TestSuite
		expectedURL string
	}{
		{
			name: "Environment variable expansion",
			suite: &models.TestSuite{
				Variables: map[string]string{
					"BASE_URL": "${TEST_VAR}",
				},
				Tests: []models.Test{
					{
						Name: "Test 1",
						Request: &models.StructuredRequest{
							URL: "${BASE_URL}/path",
						},
						Assertions: []models.Assertion{{Type: "status", Value: "200"}},
					},
				},
			},
			expectedURL: "test_value/path",
		},
		{
			name: "Literal string - no expansion needed",
			suite: &models.TestSuite{
				Variables: map[string]string{
					"VAR": "suite_value",
				},
				Tests: []models.Test{
					{
						Name:       "Test 1",
						Curl:       "curl https://example.com",
						Assertions: []models.Assertion{{Type: "status", Value: "200"}},
					},
				},
			},
			expectedURL: "",
		},
		{
			name: "Nested variable expansion",
			suite: &models.TestSuite{
				Variables: map[string]string{
					"PROTO": "https",
					"HOST":  "example.com",
					"URL":   "${PROTO}://${HOST}",
				},
				Tests: []models.Test{
					{
						Name: "Test 1",
						Request: &models.StructuredRequest{
							URL: "${URL}/api",
						},
						Assertions: []models.Assertion{{Type: "status", Value: "200"}},
					},
				},
			},
			expectedURL: "https://example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expander := NewVariableExpander()
			err := expander.ExpandVariables(tt.suite)
			if err != nil {
				t.Errorf("ExpandVariables() error = %v", err)
				return
			}

			// Verify test URL was expanded correctly (if structured request)
			if tt.expectedURL != "" && tt.suite.Tests[0].Request != nil {
				if tt.suite.Tests[0].Request.URL != tt.expectedURL {
					t.Errorf("URL = %v, want %v", tt.suite.Tests[0].Request.URL, tt.expectedURL)
				}
			}
		})
	}
}

func TestVariableExpander_ExpandString(t *testing.T) {
	expander := NewVariableExpander()
	expander.variables = map[string]string{
		"NAME":  "John",
		"AGE":   "30",
		"CITY":  "NYC",
		"EMPTY": "",
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single variable",
			input:    "Hello ${NAME}",
			expected: "Hello John",
		},
		{
			name:     "Multiple variables",
			input:    "${NAME} is ${AGE} years old",
			expected: "John is 30 years old",
		},
		{
			name:     "No variables",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "Variable at start and end",
			input:    "${NAME} lives in ${CITY}",
			expected: "John lives in NYC",
		},
		{
			name:     "Undefined variable",
			input:    "Hello ${UNDEFINED}",
			expected: "Hello ${UNDEFINED}",
		},
		{
			name:     "Empty variable",
			input:    "Value: ${EMPTY}",
			expected: "Value: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expander.expandString(tt.input)
			if result != tt.expected {
				t.Errorf("expandString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestVariableExpander_GetVariables(t *testing.T) {
	expander := NewVariableExpander()
	expander.variables = map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	vars := expander.GetVariables()
	if len(vars) != 2 {
		t.Errorf("GetVariables() returned %d variables, want 2", len(vars))
	}
	if vars["VAR1"] != "value1" {
		t.Errorf("VAR1 = %v, want value1", vars["VAR1"])
	}
	if vars["VAR2"] != "value2" {
		t.Errorf("VAR2 = %v, want value2", vars["VAR2"])
	}
}

func TestVariableExpander_ExpandTest(t *testing.T) {
	expander := NewVariableExpander()
	expander.variables = map[string]string{
		"BASE_URL": "https://api.example.com",
		"TOKEN":    "secret123",
	}

	tests := []struct {
		name string
		test models.Test
		want models.Test
	}{
		{
			name: "Expand curl command",
			test: models.Test{
				Name: "Test 1",
				Curl: "curl ${BASE_URL}/users",
			},
			want: models.Test{
				Name: "Test 1",
				Curl: "curl https://api.example.com/users",
			},
		},
		{
			name: "Expand structured request",
			test: models.Test{
				Name: "Test 2",
				Request: &models.StructuredRequest{
					Method: "GET",
					URL:    "${BASE_URL}/users",
					Headers: map[string]string{
						"Authorization": "Bearer ${TOKEN}",
					},
					Body: `{"url": "${BASE_URL}"}`,
				},
			},
			want: models.Test{
				Name: "Test 2",
				Request: &models.StructuredRequest{
					Method: "GET",
					URL:    "https://api.example.com/users",
					Headers: map[string]string{
						"Authorization": "Bearer secret123",
					},
					Body: `{"url": "https://api.example.com"}`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expander.expandTest(&tt.test)
			if err != nil {
				t.Fatalf("expandTest() error = %v", err)
			}

			if tt.test.Curl != tt.want.Curl {
				t.Errorf("Curl = %v, want %v", tt.test.Curl, tt.want.Curl)
			}

			if tt.test.Request != nil && tt.want.Request != nil {
				if tt.test.Request.URL != tt.want.Request.URL {
					t.Errorf("URL = %v, want %v", tt.test.Request.URL, tt.want.Request.URL)
				}
				if tt.test.Request.Body != tt.want.Request.Body {
					t.Errorf("Body = %v, want %v", tt.test.Request.Body, tt.want.Request.Body)
				}
				if tt.test.Request.Headers["Authorization"] != tt.want.Request.Headers["Authorization"] {
					t.Errorf("Authorization = %v, want %v",
						tt.test.Request.Headers["Authorization"],
						tt.want.Request.Headers["Authorization"])
				}
			}
		})
	}
}

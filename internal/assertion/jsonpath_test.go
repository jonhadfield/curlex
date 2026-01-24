package assertion

import (
	"testing"

	"curlex/internal/models"
)

func TestJSONPathValidator_Simple(t *testing.T) {
	validator := &JSONPathValidator{}

	jsonBody := `{
		"url": "https://httpbin.org/get",
		"headers": {
			"Host": "httpbin.org"
		},
		"data": {
			"id": 123,
			"name": "test"
		}
	}`

	tests := []struct {
		name       string
		expr       string
		shouldPass bool
	}{
		{"string equality", ".url == 'https://httpbin.org/get'", true},
		{"nested object exists", ".headers != null", true},
		{"number equality", ".data.id == 123", true},
		{"string in nested", ".data.name == 'test'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &models.TestResult{
				ResponseBody: jsonBody,
			}
			assertion := models.Assertion{
				Type:  models.AssertionJSONPath,
				Value: tt.expr,
			}

			failure := validator.Validate(result, assertion)

			if tt.shouldPass && failure != nil {
				t.Errorf("Expected to pass, but failed: %v", failure)
			}
			if !tt.shouldPass && failure == nil {
				t.Errorf("Expected to fail, but passed")
			}
		})
	}
}

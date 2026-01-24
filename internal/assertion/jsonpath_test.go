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

func TestJSONPathValidator_BooleanValues(t *testing.T) {
	validator := &JSONPathValidator{}

	jsonBody := `{
		"enabled": true,
		"disabled": false,
		"settings": {
			"active": true,
			"hidden": false
		}
	}`

	tests := []struct {
		name       string
		expr       string
		shouldPass bool
	}{
		{"boolean true equality", ".enabled == true", true},
		{"boolean false equality", ".disabled == false", true},
		{"boolean inequality", ".enabled != false", true},
		{"nested boolean true", ".settings.active == true", true},
		{"nested boolean false", ".settings.hidden == false", true},
		{"boolean true not equal false", ".enabled == false", false},
		{"boolean false not equal true", ".disabled == true", false},
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

func TestJSONPathValidator_NullValues(t *testing.T) {
	validator := &JSONPathValidator{}

	jsonBody := `{
		"value": null,
		"data": {
			"missing": null,
			"present": "value"
		}
	}`

	tests := []struct {
		name       string
		expr       string
		shouldPass bool
	}{
		{"null equality", ".value == null", true},
		{"null inequality with string", ".data.present != null", true},
		{"nested null equality", ".data.missing == null", true},
		{"non-null not equal null", ".data.present == null", false},
		{"null not equal to value", ".value != null", false},
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

func TestJSONPathValidator_NumberComparisons(t *testing.T) {
	validator := &JSONPathValidator{}

	jsonBody := `{
		"count": 42,
		"price": 19.99,
		"zero": 0,
		"negative": -5
	}`

	tests := []struct {
		name       string
		expr       string
		shouldPass bool
	}{
		{"greater than", ".count > 40", true},
		{"less than", ".count < 50", true},
		{"greater than or equal", ".count >= 42", true},
		{"less than or equal", ".count <= 42", true},
		{"decimal comparison", ".price > 19.0", true},
		{"zero comparison", ".zero == 0", true},
		{"negative comparison", ".negative < 0", true},
		{"failed greater than", ".count > 50", false},
		{"failed less than", ".count < 40", false},
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

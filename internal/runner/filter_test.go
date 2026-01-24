package runner

import (
	"testing"

	"curlex/internal/models"
)

func TestFilterTests_NoFilter(t *testing.T) {
	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "Test 1"},
			{Name: "Test 2"},
			{Name: "Test 3"},
		},
	}

	config := FilterConfig{}
	filtered := FilterTests(suite, config)

	if len(filtered) != 3 {
		t.Errorf("Expected 3 tests, got %d", len(filtered))
	}
}

func TestFilterTests_ExactName(t *testing.T) {
	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "Test 1"},
			{Name: "Test 2"},
			{Name: "Test 3"},
		},
	}

	config := FilterConfig{
		TestName: "Test 2",
	}
	filtered := FilterTests(suite, config)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 test, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Name != "Test 2" {
		t.Errorf("Expected 'Test 2', got '%s'", filtered[0].Name)
	}
}

func TestFilterTests_Pattern(t *testing.T) {
	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "API Test 1"},
			{Name: "API Test 2"},
			{Name: "UI Test 1"},
			{Name: "Integration Test"},
		},
	}

	config := FilterConfig{
		TestPattern: "^API.*",
	}
	filtered := FilterTests(suite, config)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(filtered))
	}
	for _, test := range filtered {
		if test.Name != "API Test 1" && test.Name != "API Test 2" {
			t.Errorf("Unexpected test name: %s", test.Name)
		}
	}
}

func TestFilterTests_Skip(t *testing.T) {
	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "Test 1"},
			{Name: "Test 2"},
			{Name: "Test 3"},
		},
	}

	config := FilterConfig{
		SkipTests: "Test 2",
	}
	filtered := FilterTests(suite, config)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests, got %d", len(filtered))
	}
	for _, test := range filtered {
		if test.Name == "Test 2" {
			t.Errorf("Test 2 should have been skipped")
		}
	}
}

func TestFilterTests_InvalidPattern(t *testing.T) {
	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "Test 1"},
			{Name: "Test 2"},
		},
	}

	config := FilterConfig{
		TestPattern: "[invalid",
	}
	filtered := FilterTests(suite, config)

	// Invalid pattern should return all tests
	if len(filtered) != 2 {
		t.Errorf("Expected 2 tests with invalid pattern, got %d", len(filtered))
	}
}

func TestFilterTests_CombinedPatternAndSkip(t *testing.T) {
	suite := &models.TestSuite{
		Tests: []models.Test{
			{Name: "API Test 1"},
			{Name: "API Test 2"},
			{Name: "UI Test 1"},
		},
	}

	config := FilterConfig{
		TestPattern: "^API.*",
		SkipTests:   "API Test 2",
	}
	filtered := FilterTests(suite, config)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 test, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].Name != "API Test 1" {
		t.Errorf("Expected 'API Test 1', got '%s'", filtered[0].Name)
	}
}

package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYAMLParser_Parse_Success(t *testing.T) {
	// Create temporary test file
	content := `version: "1.0"
tests:
  - name: "Test 1"
    curl: "curl https://example.com"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	suite, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if suite == nil {
		t.Fatal("Parse() returned nil suite")
	}
	if len(suite.Tests) != 1 {
		t.Errorf("Expected 1 test, got %d", len(suite.Tests))
	}
	if suite.Tests[0].Name != "Test 1" {
		t.Errorf("Test name = %v, want 'Test 1'", suite.Tests[0].Name)
	}
}

func TestYAMLParser_Parse_FileNotFound(t *testing.T) {
	parser := NewYAMLParser()
	_, err := parser.Parse("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Parse() expected error for nonexistent file")
	}
}

func TestYAMLParser_Parse_InvalidYAML(t *testing.T) {
	content := `invalid: yaml: syntax: here`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for invalid YAML")
	}
}

func TestYAMLParser_Validate_NoTests(t *testing.T) {
	content := `version: "1.0"
tests: []
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for empty test suite")
	}
	if err != nil && err.Error() != "validation failed: no tests defined in suite" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestYAMLParser_Validate_MissingName(t *testing.T) {
	content := `version: "1.0"
tests:
  - curl: "curl https://example.com"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "noname.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for missing test name")
	}
}

func TestYAMLParser_Validate_NoCurlOrRequest(t *testing.T) {
	content := `version: "1.0"
tests:
  - name: "Test 1"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "norequest.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for missing curl/request")
	}
}

func TestYAMLParser_Validate_BothCurlAndRequest(t *testing.T) {
	content := `version: "1.0"
tests:
  - name: "Test 1"
    curl: "curl https://example.com"
    request:
      method: GET
      url: "https://example.com"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "both.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for both curl and request")
	}
}

func TestYAMLParser_Validate_NoAssertions(t *testing.T) {
	content := `version: "1.0"
tests:
  - name: "Test 1"
    curl: "curl https://example.com"
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "noassertions.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for missing assertions")
	}
}

func TestYAMLParser_Validate_StructuredRequestMissingURL(t *testing.T) {
	content := `version: "1.0"
tests:
  - name: "Test 1"
    request:
      method: GET
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nourl.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for missing request URL")
	}
}

func TestYAMLParser_Validate_StructuredRequestMissingMethod(t *testing.T) {
	content := `version: "1.0"
tests:
  - name: "Test 1"
    request:
      url: "https://example.com"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nomethod.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	_, err := parser.Parse(testFile)
	if err == nil {
		t.Error("Parse() expected error for missing request method")
	}
}

func TestYAMLParser_Parse_WithVariables(t *testing.T) {
	// Set environment variable
	_ = os.Setenv("API_URL", "https://api.example.com")
	defer func() { _ = os.Unsetenv("API_URL") }()

	content := `version: "1.0"
variables:
  BASE_URL: "${API_URL}"
tests:
  - name: "Test 1"
    curl: "curl ${BASE_URL}/users"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "vars.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	suite, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check that variables were expanded in the test's curl command
	expectedCurl := "curl https://api.example.com/users"
	if suite.Tests[0].Curl != expectedCurl {
		t.Errorf("Curl = %v, want %v", suite.Tests[0].Curl, expectedCurl)
	}
}

func TestYAMLParser_Parse_WithDefaults(t *testing.T) {
	content := `version: "1.0"
defaults:
  timeout: 30s
  retries: 2
tests:
  - name: "Test 1"
    curl: "curl https://example.com"
    assertions:
      - status: 200
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "defaults.yaml")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewYAMLParser()
	suite, err := parser.Parse(testFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Defaults should be applied to tests
	if suite.Tests[0].Retries != 2 {
		t.Errorf("Test retries = %v, want 2", suite.Tests[0].Retries)
	}
}

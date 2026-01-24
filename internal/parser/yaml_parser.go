package parser

import (
	"errors"
	"fmt"
	"os"

	"curlex/internal/models"
	"gopkg.in/yaml.v3"
)

// YAMLParser parses test suite YAML files
type YAMLParser struct{}

// NewYAMLParser creates a new YAML parser instance
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// Parse reads a YAML file and returns a test suite
func (p *YAMLParser) Parse(yamlPath string) (*models.TestSuite, error) {
	// Read file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse YAML
	var suite models.TestSuite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Expand variables
	expander := NewVariableExpander()
	if err := expander.ExpandVariables(&suite); err != nil {
		return nil, fmt.Errorf("variable expansion failed: %w", err)
	}

	// Apply defaults to all tests
	ApplyDefaults(&suite)

	// Validate suite
	if err := p.validate(&suite); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &suite, nil
}

// validate performs basic validation on the test suite
func (p *YAMLParser) validate(suite *models.TestSuite) error {
	var errs []error

	if len(suite.Tests) == 0 {
		return fmt.Errorf("no tests defined in suite")
	}

	for i, test := range suite.Tests {
		// Test must have a name
		if test.Name == "" {
			errs = append(errs, fmt.Errorf("test %d: name is required", i))
		}

		// Test must have either curl or request
		if test.Curl == "" && test.Request == nil {
			testID := test.Name
			if testID == "" {
				testID = fmt.Sprintf("%d", i)
			}
			errs = append(errs, fmt.Errorf("test %s: must specify either 'curl' or 'request'", testID))
		}

		// Test cannot have both curl and request
		if test.Curl != "" && test.Request != nil {
			errs = append(errs, fmt.Errorf("test %s: cannot specify both 'curl' and 'request'", test.Name))
		}

		// Test must have at least one assertion
		if len(test.Assertions) == 0 {
			testID := test.Name
			if testID == "" {
				testID = fmt.Sprintf("%d", i)
			}
			errs = append(errs, fmt.Errorf("test %s: must have at least one assertion", testID))
		}

		// Validate structured request if present
		if test.Request != nil {
			if test.Request.URL == "" {
				errs = append(errs, fmt.Errorf("test %s: request.url is required", test.Name))
			}
			if test.Request.Method == "" {
				errs = append(errs, fmt.Errorf("test %s: request.method is required", test.Name))
			}
		}
	}

	return errors.Join(errs...)
}

package models

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// AssertionType represents the type of assertion
type AssertionType string

const (
	AssertionStatus       AssertionType = "status"
	AssertionBody         AssertionType = "body"
	AssertionBodyContains AssertionType = "body_contains"
	AssertionJSONPath     AssertionType = "json_path"
	AssertionHeader       AssertionType = "header"
	AssertionResponseTime AssertionType = "response_time"
)

// Assertion represents a single test assertion
type Assertion struct {
	Type  AssertionType
	Value string
}

// UnmarshalYAML implements custom YAML unmarshaling for flexible assertion syntax
// Supports both formats:
// - status: 200
// - json_path: ".data.id == 1"
func (a *Assertion) UnmarshalYAML(value *yaml.Node) error {
	// Parse as map to get the assertion type and value
	var assertionMap map[string]string
	if err := value.Decode(&assertionMap); err != nil {
		return fmt.Errorf("failed to decode assertion: %w", err)
	}

	// Should have exactly one key
	if len(assertionMap) != 1 {
		return fmt.Errorf("assertion must have exactly one key-value pair, got %d", len(assertionMap))
	}

	// Extract type and value
	for key, val := range assertionMap {
		assertionType := strings.TrimSpace(key)

		// Validate assertion type
		switch assertionType {
		case "status":
			a.Type = AssertionStatus
		case "body":
			a.Type = AssertionBody
		case "body_contains":
			a.Type = AssertionBodyContains
		case "json_path":
			a.Type = AssertionJSONPath
		case "header":
			a.Type = AssertionHeader
		case "response_time":
			a.Type = AssertionResponseTime
		default:
			return fmt.Errorf("unknown assertion type: %s", assertionType)
		}

		a.Value = val
	}

	return nil
}

// String returns a human-readable representation of the assertion
func (a Assertion) String() string {
	return fmt.Sprintf("%s: %s", a.Type, a.Value)
}

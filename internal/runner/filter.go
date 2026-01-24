package runner

import (
	"regexp"

	"curlex/internal/models"
)

// FilterConfig holds test filtering configuration
type FilterConfig struct {
	TestName    string // Exact test name to run
	TestPattern string // Regex pattern for test names
	SkipTests   string // Test name to skip
}

// FilterTests filters the test suite based on configuration
func FilterTests(suite *models.TestSuite, config FilterConfig) []models.Test {
	if config.TestName == "" && config.TestPattern == "" && config.SkipTests == "" {
		// No filtering - return all tests
		return suite.Tests
	}

	var filtered []models.Test

	// Compile regex pattern if provided
	var pattern *regexp.Regexp
	if config.TestPattern != "" {
		var err error
		pattern, err = regexp.Compile(config.TestPattern)
		if err != nil {
			// Invalid pattern - return all tests
			return suite.Tests
		}
	}

	for _, test := range suite.Tests {
		// Skip if test name matches skip pattern
		if config.SkipTests != "" && test.Name == config.SkipTests {
			continue
		}

		// Include test if it matches the filter
		include := false

		if config.TestName != "" {
			// Exact name match
			include = test.Name == config.TestName
		} else if pattern != nil {
			// Regex pattern match
			include = pattern.MatchString(test.Name)
		} else {
			// No specific filter, just applying skip logic
			include = true
		}

		if include {
			filtered = append(filtered, test)
		}
	}

	return filtered
}

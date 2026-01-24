package output

import (
	"encoding/xml"
	"fmt"
	"strings"

	"curlex/internal/models"
)

// JUnitFormatter formats test results as JUnit XML
type JUnitFormatter struct{}

// NewJUnitFormatter creates a new JUnit XML formatter
func NewJUnitFormatter() *JUnitFormatter {
	return &JUnitFormatter{}
}

// JUnitTestSuites is the root element
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite represents a test suite
type JUnitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Time     float64         `xml:"time,attr"`
	Cases    []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase represents a single test case
type JUnitTestCase struct {
	Name      string          `xml:"name,attr"`
	Classname string          `xml:"classname,attr"`
	Time      float64         `xml:"time,attr"`
	Failure   *JUnitFailure   `xml:"failure,omitempty"`
	Error     *JUnitError     `xml:"error,omitempty"`
	SystemOut string          `xml:"system-out,omitempty"`
}

// JUnitFailure represents a test failure
type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// JUnitError represents a test error
type JUnitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// Format converts suite results to JUnit XML
func (f *JUnitFormatter) Format(suiteResult *models.SuiteResult) string {
	suite := JUnitTestSuite{
		Name:     "curlex",
		Tests:    suiteResult.TotalTests,
		Failures: suiteResult.FailedTests,
		Errors:   0,
		Time:     suiteResult.TotalTime.Seconds(),
		Cases:    make([]JUnitTestCase, 0, len(suiteResult.Results)),
	}

	for _, result := range suiteResult.Results {
		testCase := JUnitTestCase{
			Name:      result.Test.Name,
			Classname: "curlex.tests",
			Time:      result.ResponseTime.Seconds(),
		}

		// Add system output (request/response details)
		var sysOut strings.Builder
		if result.PreparedRequest != nil {
			sysOut.WriteString(fmt.Sprintf("Request: %s %s\n", result.PreparedRequest.Method, result.PreparedRequest.URL))
		}
		sysOut.WriteString(fmt.Sprintf("Status: %d\n", result.StatusCode))
		sysOut.WriteString(fmt.Sprintf("Response Time: %dms\n", result.ResponseTime.Milliseconds()))
		testCase.SystemOut = sysOut.String()

		// Add failure if test failed
		if !result.Success {
			if result.Error != nil {
				// Error during execution
				testCase.Error = &JUnitError{
					Message: "Test execution error",
					Type:    "ExecutionError",
					Content: result.Error.Error(),
				}
				suite.Errors++
			} else if len(result.Failures) > 0 {
				// Assertion failures
				var failureMsg strings.Builder
				for i, failure := range result.Failures {
					if i > 0 {
						failureMsg.WriteString("\n")
					}
					failureMsg.WriteString(failure.String())
				}

				testCase.Failure = &JUnitFailure{
					Message: fmt.Sprintf("%d assertion(s) failed", len(result.Failures)),
					Type:    "AssertionFailure",
					Content: failureMsg.String(),
				}
			}
		}

		suite.Cases = append(suite.Cases, testCase)
	}

	testSuites := JUnitTestSuites{
		Suites: []JUnitTestSuite{suite},
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(testSuites, "", "  ")
	if err != nil {
		return `<?xml version="1.0" encoding="UTF-8"?><error>Failed to generate JUnit XML</error>`
	}

	return xml.Header + string(output) + "\n"
}

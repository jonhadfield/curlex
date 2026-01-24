package assertion

import (
	"curlex/internal/models"
)

// Engine validates assertions against test results
type Engine struct {
	validators map[models.AssertionType]Validator
}

// Validator interface for assertion validation
type Validator interface {
	Validate(result *models.TestResult, assertion models.Assertion) *models.AssertionFailure
}

// NewEngine creates a new assertion engine with all validators
func NewEngine() *Engine {
	return &Engine{
		validators: map[models.AssertionType]Validator{
			models.AssertionStatus:       &StatusValidator{},
			models.AssertionBody:         &BodyValidator{},
			models.AssertionBodyContains: &BodyContainsValidator{},
			models.AssertionJSONPath:     &JSONPathValidator{},
			models.AssertionHeader:       &HeaderValidator{},
			models.AssertionResponseTime: &ResponseTimeValidator{},
		},
	}
}

// Validate checks all assertions against the result
func (e *Engine) Validate(result *models.TestResult, assertions []models.Assertion) []models.AssertionFailure {
	var failures []models.AssertionFailure

	for _, assertion := range assertions {
		validator, ok := e.validators[assertion.Type]
		if !ok {
			failures = append(failures, models.AssertionFailure{
				Type:    assertion.Type,
				Message: "unsupported assertion type: " + string(assertion.Type),
			})
			continue
		}

		if failure := validator.Validate(result, assertion); failure != nil {
			failures = append(failures, *failure)
		}
	}

	return failures
}

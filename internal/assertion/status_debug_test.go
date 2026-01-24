package assertion

import (
	"fmt"
	"testing"
)

func TestDebugExtractComparison(t *testing.T) {
	validator := &StatusValidator{}

	expr := ">= 200"

	match := validator.extractComparisonRe(expr, statusPatternGTE)
	fmt.Printf("expr: %q, pattern: statusPatternGTE\n", expr)
	fmt.Printf("match: %v\n", match)

	if match != nil {
		fmt.Printf("match[0]: %q\n", match[0])
		fmt.Printf("match[1]: %q\n", match[1])
	}
}

func TestDebugEvaluateSingle(t *testing.T) {
	validator := &StatusValidator{}

	result := validator.evaluateSingleExpression(">= 200", 201)
	fmt.Printf("evaluateSingleExpression('>= 200', 201) = %v\n", result)

	result2 := validator.evaluateSingleExpression("< 300", 201)
	fmt.Printf("evaluateSingleExpression('< 300', 201) = %v\n", result2)
}

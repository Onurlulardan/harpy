package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

// CompileExactCount compiles the separately budgeted root-set count statement.
func (compiler Compiler) CompileExactCount(
	plan resolved.AuthorizedPlan,
) (ExactCountStatement, error) {
	if compiler.Validate() != nil || plan.Validate() != nil {
		return ExactCountStatement{}, ErrContract
	}
	input, err := compiler.extractRecord(plan, api.CountModeExact)
	if err != nil {
		return ExactCountStatement{}, err
	}
	return compileExactCount(input)
}

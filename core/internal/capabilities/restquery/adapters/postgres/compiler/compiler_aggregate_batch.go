package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

// CompileAggregateBatch atomically mints the aggregate statement.
func (compiler Compiler) CompileAggregateBatch(
	plan resolved.AuthorizedPlan,
) (AggregateBatch, error) {
	if compiler.Validate() != nil || plan.Validate() != nil {
		return AggregateBatch{}, ErrContract
	}
	binding, err := aggregatePlanBinding(plan)
	if err != nil {
		return AggregateBatch{}, err
	}
	input, err := compiler.extractAggregate(plan)
	if err != nil || input.binding != binding {
		return AggregateBatch{}, ErrContract
	}
	statement, err := compileAggregate(input)
	if err != nil {
		return AggregateBatch{}, err
	}
	return newAggregateBatch(statement, binding)
}

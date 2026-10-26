package compiler

import (
	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

// CompileV1Page turns a sealed predecessor tuple directly into keyset SQL.
func (compiler Compiler) CompileV1Page(
	plan resolved.AuthorizedPlan,
	ordinal v1query.PredecessorOrdinal,
	predecessor PredecessorStatement,
	tuple PredecessorTuple,
) (RecordStatement, error) {
	if compiler.Validate() != nil || predecessor.ValidatePlan(plan, ordinal) != nil ||
		tuple.validateFor(predecessor) != nil {
		return RecordStatement{}, ErrContract
	}
	base, err := v1BaseBinding(plan)
	if err != nil {
		return RecordStatement{}, ErrContract
	}
	input, err := compiler.extractRecord(plan, api.CountModeExact)
	if err != nil || len(input.last) != 0 || input.binding != base {
		return RecordStatement{}, ErrContract
	}
	input.last = append(input.last, tuple.values...)
	input.binding, err = v1PageBinding(
		base, predecessor.stateSeal, tuple.stateSeal,
	)
	if err != nil {
		return RecordStatement{}, ErrContract
	}
	statement, err := compileRecord(input)
	if err != nil || statement.validateV1Plan(
		plan, ordinal, predecessor, tuple,
	) != nil {
		return RecordStatement{}, ErrContract
	}
	return statement, nil
}

func (statement RecordStatement) validateV1Plan(
	plan resolved.AuthorizedPlan,
	ordinal v1query.PredecessorOrdinal,
	predecessor PredecessorStatement,
	tuple PredecessorTuple,
) error {
	if statement.Validate() != nil || predecessor.ValidatePlan(plan, ordinal) != nil ||
		tuple.validateFor(predecessor) != nil {
		return ErrContract
	}
	base, err := v1BaseBinding(plan)
	if err != nil {
		return ErrContract
	}
	expected, err := v1PageBinding(base, predecessor.stateSeal, tuple.stateSeal)
	if err != nil || expected != statement.binding ||
		statement.countMode != api.CountModeExact {
		return ErrContract
	}
	return nil
}

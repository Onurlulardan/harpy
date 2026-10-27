package compiler

import (
	"strings"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

// CompileV1Predecessor mints the only OFFSET-bearing REST Query statement.
func (compiler Compiler) CompileV1Predecessor(
	plan resolved.AuthorizedPlan, ordinal v1query.PredecessorOrdinal,
) (PredecessorStatement, error) {
	base, predecessorIndex, ordinalErr := v1PlanOrdinal(plan, ordinal)
	if compiler.Validate() != nil || ordinalErr != nil {
		return PredecessorStatement{}, ErrContract
	}
	input, err := compiler.extractRecord(plan, api.CountModeExact)
	if err != nil || len(input.last) != 0 {
		return PredecessorStatement{}, ErrContract
	}
	if input.binding != base {
		return PredecessorStatement{}, ErrContract
	}
	columns, err := v1PredecessorColumns(input.query)
	if err != nil {
		return PredecessorStatement{}, err
	}
	binder := newBinder()
	sql, err := compileV1PredecessorSQL(input, columns, binder, predecessorIndex)
	if err != nil {
		return PredecessorStatement{}, err
	}
	binding, err := v1PredecessorBinding(base, predecessorIndex)
	if err != nil {
		return PredecessorStatement{}, err
	}
	return newPredecessorStatement(
		sql, binder.arguments, resultColumns(columns), predecessorIndex,
		planner.StatementClassV1PredecessorSeek, binding,
	)
}

func compileV1PredecessorSQL(
	input recordInput, columns []recordColumn, binder *binder, index int64,
) (string, error) {
	var builder strings.Builder
	builder.WriteString("SELECT ")
	if appendProjectionList(&builder, columns) != nil {
		return "", ErrContract
	}
	builder.WriteString(" FROM ")
	if appendResourceStorage(&builder, input.storage) != nil {
		return "", ErrContract
	}
	conditions, err := rootConditions(input.query, input.visibility, binder)
	if err != nil {
		return "", err
	}
	if len(conditions) != 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}
	if appendRecordOrder(&builder, input.query, binder) != nil {
		return "", ErrContract
	}
	limit, limitErr := binder.bind(1)
	offset, offsetErr := binder.bind(index)
	if limitErr != nil || offsetErr != nil {
		return "", ErrContract
	}
	builder.WriteString(" LIMIT ")
	builder.WriteString(limit)
	builder.WriteString(" OFFSET ")
	builder.WriteString(offset)
	return builder.String(), nil
}

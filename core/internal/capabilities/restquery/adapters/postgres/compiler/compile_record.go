package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func compileRecord(input recordInput) (RecordStatement, error) {
	if input.validate() != nil {
		return RecordStatement{}, ErrContract
	}
	binder := newBinder()
	columns, err := recordColumns(input.query)
	if err != nil {
		return RecordStatement{}, err
	}
	var builder strings.Builder
	builder.WriteString("SELECT ")
	if err := appendProjectionList(&builder, columns); err != nil {
		return RecordStatement{}, err
	}
	builder.WriteString(" FROM ")
	if err := appendResourceStorage(&builder, input.storage); err != nil {
		return RecordStatement{}, err
	}
	conditions, err := recordConditions(input, binder)
	if err != nil {
		return RecordStatement{}, err
	}
	if len(conditions) != 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}
	if err := appendRecordOrder(&builder, input.query, binder); err != nil {
		return RecordStatement{}, err
	}
	page, err := input.query.Page()
	if err != nil {
		return RecordStatement{}, ErrContract
	}
	requested, err := page.Size()
	if err != nil {
		return RecordStatement{}, ErrContract
	}
	limit, err := binder.bind(requested + 1)
	if err != nil {
		return RecordStatement{}, err
	}
	builder.WriteString(" LIMIT ")
	builder.WriteString(limit)
	count, err := input.query.Count()
	if err != nil {
		return RecordStatement{}, ErrContract
	}
	return newRecordStatement(
		builder.String(), binder.arguments, resultColumns(columns), requested,
		planner.StatementClassRecord, count, input.context, input.binding,
	)
}

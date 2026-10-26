package compiler

import "strings"

func compileAggregate(input aggregateInput) (AggregateStatement, error) {
	if input.validate() != nil {
		return AggregateStatement{}, ErrContract
	}
	groups, measures, err := aggregateColumns(input)
	if err != nil {
		return AggregateStatement{}, err
	}
	binder := newBinder()
	var builder strings.Builder
	builder.WriteString("WITH ")
	if err := appendAggregateRawCTE(&builder, input, binder); err != nil {
		return AggregateStatement{}, err
	}
	builder.WriteString(", ")
	if err := appendAggregateTypedCTE(&builder, input, binder); err != nil {
		return AggregateStatement{}, err
	}
	builder.WriteString(", ")
	if err := appendAggregateGuardCTE(&builder, len(input.measures)); err != nil {
		return AggregateStatement{}, err
	}
	builder.WriteByte(' ')
	if err := appendAggregateSelect(&builder, input, binder); err != nil {
		return AggregateStatement{}, err
	}
	limit, err := input.query.Limit()
	if err != nil {
		return AggregateStatement{}, ErrContract
	}
	return newAggregateStatement(
		builder.String(), binder.arguments, groups, measures, limit, input.binding,
	)
}

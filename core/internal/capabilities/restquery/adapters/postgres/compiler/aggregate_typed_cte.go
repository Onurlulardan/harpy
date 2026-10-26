package compiler

import "strings"

func appendAggregateTypedCTE(
	builder *strings.Builder, input aggregateInput, binder *binder,
) error {
	if builder == nil || binder == nil || input.validate() != nil {
		return ErrContract
	}
	projections := make([]string, 0, len(input.groups)+len(input.measures))
	for index := range input.groups {
		alias, err := aggregateGroupAlias(index)
		if err != nil {
			return err
		}
		projections = append(projections, alias)
	}
	for index, measure := range input.measures {
		expression, err := aggregateTypedMeasure(measure, index, binder)
		alias, aliasErr := aggregateMeasureAlias(index)
		if err != nil || aliasErr != nil {
			return ErrContract
		}
		projections = append(projections, expression+" AS "+alias)
	}
	builder.WriteString(aggregateTypedCTE)
	builder.WriteString(" AS MATERIALIZED (SELECT ")
	builder.WriteString(strings.Join(projections, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(aggregateRawCTE)
	builder.WriteByte(')')
	return nil
}

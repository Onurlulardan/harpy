package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func appendAggregateRawCTE(
	builder *strings.Builder,
	input aggregateInput,
	binder *binder,
) error {
	if builder == nil || binder == nil || input.validate() != nil {
		return ErrContract
	}
	projections := make([]string, 0, len(input.groups)+len(input.measures))
	groupExpressions := make([]string, len(input.groups))
	for index, group := range input.groups {
		expression, err := exactExpression(group.field)
		alias, aliasErr := aggregateGroupAlias(index)
		if err != nil || aliasErr != nil {
			return ErrContract
		}
		groupExpressions[index] = expression
		projections = append(projections, expression+" AS "+alias)
	}
	for index, measure := range input.measures {
		expression, err := aggregateRawMeasure(measure, binder)
		alias, aliasErr := aggregateMeasureAlias(index)
		if err != nil || aliasErr != nil {
			return ErrContract
		}
		projections = append(projections, expression+" AS "+alias)
		if measure.function == api.AggregateFunctionAvg {
			count, countErr := aggregateAverageCount(measure)
			countAlias, aliasErr := aggregateAverageCountAlias(index)
			if countErr != nil || aliasErr != nil {
				return ErrContract
			}
			projections = append(projections, count+" AS "+countAlias)
		}
	}
	builder.WriteString(aggregateRawCTE)
	builder.WriteString(" AS MATERIALIZED (SELECT ")
	builder.WriteString(strings.Join(projections, ", "))
	builder.WriteString(" FROM ")
	if appendIdentifier(builder, input.table) != nil {
		return ErrContract
	}
	conditions, err := aggregateConditions(input.query, input.visibility, binder)
	if err != nil {
		return err
	}
	if len(conditions) != 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}
	if len(groupExpressions) != 0 {
		builder.WriteString(" GROUP BY ")
		builder.WriteString(strings.Join(groupExpressions, ", "))
	}
	builder.WriteByte(')')
	return nil
}

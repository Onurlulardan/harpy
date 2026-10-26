package compiler

import "strings"

func appendAggregateSelect(
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
	for index := range input.measures {
		alias, err := aggregateMeasureAlias(index)
		if err != nil {
			return err
		}
		projections = append(projections, alias)
	}
	guard, err := aggregateGuardCondition(len(input.measures))
	if err != nil {
		return err
	}
	conditions := []string{parenthesize(guard)}
	having, present, err := input.query.Having()
	if err != nil {
		return ErrContract
	}
	if present {
		compiled, err := compileAggregateHaving(having, input, binder)
		if err != nil {
			return err
		}
		conditions = append(conditions, parenthesize(compiled))
	}
	builder.WriteString("SELECT ")
	builder.WriteString(strings.Join(projections, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(aggregateTypedCTE)
	builder.WriteString(" CROSS JOIN ")
	builder.WriteString(aggregateGuardCTE)
	builder.WriteString(" WHERE ")
	builder.WriteString(strings.Join(conditions, " AND "))
	if err := appendAggregateOrder(builder, input, binder); err != nil {
		return err
	}
	limit, err := input.query.Limit()
	if err != nil {
		return ErrContract
	}
	placeholder, err := binder.bind(limit)
	if err != nil {
		return err
	}
	builder.WriteString(" LIMIT ")
	builder.WriteString(placeholder)
	return nil
}

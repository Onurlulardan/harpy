package compiler

import "strings"

func appendAggregateGuardCTE(
	builder *strings.Builder, measureCount int,
) error {
	if builder == nil || measureCount < 0 || measureCount > 8 {
		return ErrContract
	}
	checks := []string{"pg_catalog.count(*) AS " + aggregateGuardKey}
	for ordinal := 0; ordinal < measureCount; ordinal++ {
		measure, measureErr := aggregateMeasureAlias(ordinal)
		guard, guardErr := aggregateGuardAlias(ordinal)
		if measureErr != nil || guardErr != nil {
			return ErrContract
		}
		checks = append(checks, "pg_catalog.count("+measure+") AS "+guard)
	}
	builder.WriteString(aggregateGuardCTE)
	builder.WriteString(" AS MATERIALIZED (SELECT ")
	builder.WriteString(strings.Join(checks, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(aggregateTypedCTE)
	builder.WriteByte(')')
	return nil
}

func aggregateGuardCondition(measureCount int) (string, error) {
	if measureCount < 0 || measureCount > 8 {
		return "", ErrContract
	}
	if measureCount == 0 {
		return aggregateGuardKey + " >= 0", nil
	}
	checks := make([]string, measureCount)
	for ordinal := 0; ordinal < measureCount; ordinal++ {
		guard, err := aggregateGuardAlias(ordinal)
		if err != nil {
			return "", err
		}
		checks[ordinal] = guard + " <= " + aggregateGuardKey
	}
	return strings.Join(checks, " AND "), nil
}

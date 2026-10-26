package compiler

func aggregateRoundEven(
	rawAlias string, divisor string, ordinal int, scale int, binder *binder,
) (string, error) {
	if rawAlias == "" || divisor == "" || scale < 0 || scale > 1000 || binder == nil {
		return "", ErrContract
	}
	scaleValue, err := binder.bind(scale)
	if err != nil {
		return "", err
	}
	scaled, scaledErr := aggregateRoundAlias(ordinal)
	divisorAlias, divisorErr := aggregateDivisorAlias(ordinal)
	quotient, quotientErr := aggregateQuotientAlias(ordinal)
	if scaledErr != nil || divisorErr != nil || quotientErr != nil {
		return "", ErrContract
	}
	factor := "pg_catalog.power(CAST(10 AS pg_catalog.numeric), " + scaleValue + ")"
	remainder := "pg_catalog.abs(" + scaled + " - " + quotient + " * " + divisorAlias + ")"
	doubled := "2 * " + remainder
	delta := "CASE WHEN " + doubled + " < pg_catalog.abs(" + divisorAlias +
		") THEN 0 WHEN " + doubled + " > pg_catalog.abs(" + divisorAlias +
		") THEN pg_catalog.sign(" + scaled + ") WHEN pg_catalog.mod(pg_catalog.abs(" +
		quotient + "), 2) = 0 THEN 0 ELSE pg_catalog.sign(" + scaled + ") END"
	inner := "SELECT " + scaled + ", " + divisorAlias +
		", pg_catalog.div(" + scaled + ", " + divisorAlias + ") AS " + quotient +
		" FROM (SELECT " + rawAlias + " * " + factor + " AS " + scaled +
		", " + divisor + " AS " + divisorAlias + ") AS \"__fc_scaled\""
	return "(SELECT CASE WHEN " + divisorAlias + " = 0 OR " + scaled +
		" IS NULL THEN NULL ELSE (" + quotient + " + " + delta + ") / " + factor +
		" END FROM (" + inner + ") AS \"__fc_quotient\")", nil
}

package recordkernel

func applyCalculations(
	catalog Catalog,
	record *Record,
	budget *instructionBudget,
) error {
	for _, rule := range catalog.calculations {
		field := catalog.fieldsByID[rule.fieldID]
		value, err := evaluateExpression(
			rule.expression, *record, field.Decimal, budget,
		)
		if err != nil {
			return err
		}
		coerced, err := coerceExpressionValue(field, value)
		if err != nil {
			return err
		}
		record.values[rule.fieldID] = coerced
	}
	return nil
}

package recordkernel

func applyDefaults(
	catalog Catalog,
	record *Record,
	budget *instructionBudget,
) error {
	for _, rule := range catalog.defaults {
		if _, present := record.values[rule.fieldID]; present {
			continue
		}
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

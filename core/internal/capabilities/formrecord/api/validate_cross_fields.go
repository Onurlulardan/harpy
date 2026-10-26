package recordkernel

func validateCrossFields(
	catalog Catalog,
	record Record,
	budget *instructionBudget,
) ([]ValidationIssue, error) {
	issues := make([]ValidationIssue, 0)
	target := maximumCatalogDecimalProfile(catalog)
	for _, rule := range catalog.crossFieldValidations {
		value, err := evaluateExpression(rule.predicate, record, target, budget)
		if err != nil {
			return nil, err
		}
		valid := false
		if !value.null {
			valid, err = valueBoolean(value)
			if err != nil {
				return nil, err
			}
		}
		if !valid {
			issues = append(issues, ValidationIssue{
				Code: ValidationCrossField, RuleID: rule.id,
			})
		}
	}
	return issues, nil
}

func maximumCatalogDecimalProfile(catalog Catalog) *DecimalProfile {
	result := DecimalProfile{}
	for _, field := range catalog.fields {
		if field.Decimal != nil {
			if field.Decimal.Precision > result.Precision {
				result.Precision = field.Decimal.Precision
			}
			if field.Decimal.Scale > result.Scale {
				result.Scale = field.Decimal.Scale
			}
		}
	}
	if result.Precision == 0 {
		return nil
	}
	return &result
}

package recordkernel

func defaultDocuments(values []defaultValue) ([]catalogAssignment, error) {
	result := make([]catalogAssignment, len(values))
	for index, value := range values {
		expression, err := expressionDocument(value.expression)
		if err != nil {
			return nil, err
		}
		result[index] = catalogAssignment{
			FieldID: value.fieldID.String(), Expression: expression,
		}
	}
	return result, nil
}

func calculationDocuments(values []calculation) ([]catalogAssignment, error) {
	result := make([]catalogAssignment, len(values))
	for index, value := range values {
		expression, err := expressionDocument(value.expression)
		if err != nil {
			return nil, err
		}
		result[index] = catalogAssignment{
			FieldID: value.fieldID.String(), Expression: expression,
		}
	}
	return result, nil
}

func crossRuleDocuments(values []crossFieldRule) ([]catalogCrossRule, error) {
	result := make([]catalogCrossRule, len(values))
	for index, value := range values {
		predicate, err := expressionDocument(value.predicate)
		if err != nil {
			return nil, err
		}
		result[index] = catalogCrossRule{
			ID: value.id.String(), Code: value.code, Predicate: predicate,
		}
	}
	return result, nil
}

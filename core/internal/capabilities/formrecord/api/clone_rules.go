package recordkernel

func cloneRecordRules(source RecordRules) RecordRules {
	result := source
	result.Fields = append([]RecordFieldRule(nil), source.Fields...)
	result.WorkflowActions = append([]string(nil), source.WorkflowActions...)
	result.Defaults = make([]DefaultRule, len(source.Defaults))
	for index, rule := range source.Defaults {
		result.Defaults[index] = DefaultRule{
			FieldID: rule.FieldID, Expression: cloneTypedExpression(rule.Expression),
		}
	}
	result.CrossFieldValidations = make([]CrossFieldValidation,
		len(source.CrossFieldValidations))
	for index, rule := range source.CrossFieldValidations {
		result.CrossFieldValidations[index] = CrossFieldValidation{
			ID: rule.ID, Code: rule.Code,
			Predicate: cloneTypedExpression(rule.Predicate),
		}
	}
	return result
}

func cloneTypedExpression(source TypedExpression) TypedExpression {
	result := source
	if source.FieldID != nil {
		fieldID := *source.FieldID
		result.FieldID = &fieldID
	}
	if source.Operator != nil {
		operator := *source.Operator
		result.Operator = &operator
	}
	if source.Function != nil {
		function := *source.Function
		result.Function = &function
	}
	result.Value = append([]byte(nil), source.Value...)
	if source.Arguments != nil {
		result.Arguments = make([]TypedExpression, len(source.Arguments))
		for index := range source.Arguments {
			result.Arguments[index] = cloneTypedExpression(source.Arguments[index])
		}
	}
	return result
}

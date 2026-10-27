package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileComparisonPredicate(
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, error) {
	if operator == api.OperatorBetween {
		if len(values) != 2 {
			return "", ErrContract
		}
		expression, err := comparisonExpression(field, operator, binder)
		if err != nil {
			return "", err
		}
		lower, err := bindComparisonValue(field, operator, values[0], binder)
		if err != nil {
			return "", err
		}
		upper, err := bindComparisonValue(field, operator, values[1], binder)
		if err != nil {
			return "", err
		}
		return expression + " BETWEEN " + lower + " AND " + upper, nil
	}
	operatorSQL, valid := scalarOperatorSQL(operator)
	if !valid || len(values) != 1 {
		return "", ErrContract
	}
	expression, err := comparisonExpression(field, operator, binder)
	if err != nil {
		return "", err
	}
	placeholder, err := bindComparisonValue(field, operator, values[0], binder)
	if err != nil {
		return "", err
	}
	return expression + " " + operatorSQL + " " + placeholder, nil
}

func comparisonExpression(
	field api.FieldDefinition, operator api.Operator, binder *binder,
) (string, error) {
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", err
	}
	if kind == api.LogicalTypeEnum && isOrderOperator(operator) {
		return enumOrderExpression(field, binder)
	}
	return exactExpression(field)
}

func bindComparisonValue(
	field api.FieldDefinition,
	operator api.Operator,
	value resolved.Value,
	binder *binder,
) (string, error) {
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", err
	}
	if kind == api.LogicalTypeEnum && isOrderOperator(operator) {
		return bindEnumOrdinal(field, value, binder)
	}
	return bindFieldValue(field, value, binder)
}

func scalarOperatorSQL(operator api.Operator) (string, bool) {
	switch operator {
	case api.OperatorEq:
		return "=", true
	case api.OperatorNe:
		return "<>", true
	case api.OperatorGT:
		return ">", true
	case api.OperatorGTE:
		return ">=", true
	case api.OperatorLT:
		return "<", true
	case api.OperatorLTE:
		return "<=", true
	default:
		return "", false
	}
}

func isOrderOperator(operator api.Operator) bool {
	return operator == api.OperatorGT || operator == api.OperatorGTE ||
		operator == api.OperatorLT || operator == api.OperatorLTE ||
		operator == api.OperatorBetween
}

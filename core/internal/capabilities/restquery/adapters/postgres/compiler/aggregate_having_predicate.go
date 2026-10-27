package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileAggregateHavingPredicate(
	predicate resolved.HavingPredicate,
	input aggregateInput,
	binder *binder,
) (string, error) {
	reference, referenceErr := predicate.Reference()
	operator, operatorErr := predicate.Operator()
	values, valuesErr := predicate.Values()
	if referenceErr != nil || operatorErr != nil || valuesErr != nil {
		return "", ErrContract
	}
	target, err := aggregateReferenceTarget(input, reference)
	if err != nil {
		return "", err
	}
	ordered := isOrderOperator(operator)
	expression, err := aggregateTargetExpression(target, ordered, binder)
	if err != nil {
		return "", err
	}
	if operator == api.OperatorIsNull || operator == api.OperatorIsNotNull {
		if len(values) != 0 {
			return "", ErrContract
		}
		if operator == api.OperatorIsNull {
			return expression + " IS NULL", nil
		}
		return expression + " IS NOT NULL", nil
	}
	if operator == api.OperatorIn || operator == api.OperatorNotIn {
		return compileAggregateHavingList(expression, target, operator, values, binder)
	}
	if operator == api.OperatorBetween {
		if len(values) != 2 {
			return "", ErrContract
		}
		lower, lowerErr := bindAggregateTargetValue(target, values[0], true, binder)
		upper, upperErr := bindAggregateTargetValue(target, values[1], true, binder)
		if lowerErr != nil || upperErr != nil {
			return "", ErrContract
		}
		return expression + " BETWEEN " + lower + " AND " + upper, nil
	}
	operatorSQL, valid := scalarOperatorSQL(operator)
	if !valid || len(values) != 1 {
		return "", ErrContract
	}
	placeholder, err := bindAggregateTargetValue(target, values[0], ordered, binder)
	if err != nil {
		return "", err
	}
	return expression + " " + operatorSQL + " " + placeholder, nil
}

func compileAggregateHavingList(
	expression string, target aggregateTarget, operator api.Operator,
	values []resolved.Value, binder *binder,
) (string, error) {
	if len(values) < 1 || len(values) > 100 {
		return "", ErrContract
	}
	placeholders := make([]string, len(values))
	for index, value := range values {
		placeholder, err := bindAggregateTargetValue(target, value, false, binder)
		if err != nil {
			return "", err
		}
		placeholders[index] = placeholder
	}
	keyword := " IN ("
	if operator == api.OperatorNotIn {
		keyword = " NOT IN ("
	}
	return expression + keyword + strings.Join(placeholders, ", ") + ")", nil
}

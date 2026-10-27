package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compilePredicate(
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, error) {
	if field.Validate() != nil || binder == nil {
		return "", ErrContract
	}
	if statement, handled, err := compileJSONBPredicate(
		field, operator, values, binder,
	); handled || err != nil {
		return statement, err
	}
	if operator == api.OperatorIsNull || operator == api.OperatorIsNotNull {
		if len(values) != 0 {
			return "", ErrContract
		}
		expression, err := baseExpression(field)
		if err != nil {
			return "", err
		}
		if operator == api.OperatorIsNull {
			return expression + " IS NULL", nil
		}
		return expression + " IS NOT NULL", nil
	}
	if operator == api.OperatorContains || operator == api.OperatorStartsWith ||
		operator == api.OperatorEndsWith {
		if len(values) != 1 {
			return "", ErrContract
		}
		return compilePattern(field, operator, values[0], binder)
	}
	if operator == api.OperatorContainsAny || operator == api.OperatorContainsAll {
		return compileCollectionPredicate(field, operator, values, binder)
	}
	if operator == api.OperatorIn || operator == api.OperatorNotIn {
		return compileListPredicate(field, operator, values, binder)
	}
	return compileComparisonPredicate(field, operator, values, binder)
}

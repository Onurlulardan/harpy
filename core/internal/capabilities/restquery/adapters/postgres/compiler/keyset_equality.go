package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func keysetEquality(
	sortValue resolved.Sort, last cursor.LastValue, binder *binder,
) (string, error) {
	field, err := sortValue.Field()
	if err != nil {
		return "", ErrContract
	}
	expression, err := exactExpression(field)
	if err != nil {
		return "", err
	}
	isNull, err := last.IsNull()
	if err != nil {
		return "", ErrContract
	}
	if isNull {
		return expression + " IS NULL", nil
	}
	value, present, err := last.Value()
	if err != nil || !present {
		return "", ErrContract
	}
	placeholder, err := bindFieldValue(field, value, binder)
	if err != nil {
		return "", err
	}
	return expression + " = " + placeholder, nil
}

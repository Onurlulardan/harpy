package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileListPredicate(
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, error) {
	if len(values) == 0 || len(values) > 100 {
		return "", ErrContract
	}
	expression, err := exactExpression(field)
	if err != nil {
		return "", err
	}
	placeholders := make([]string, len(values))
	for index, value := range values {
		placeholder, valueErr := bindFieldValue(field, value, binder)
		if valueErr != nil {
			return "", valueErr
		}
		placeholders[index] = placeholder
	}
	keyword := " IN ("
	if operator == api.OperatorNotIn {
		keyword = " NOT IN ("
	} else if operator != api.OperatorIn {
		return "", ErrContract
	}
	return expression + keyword + strings.Join(placeholders, ", ") + ")", nil
}

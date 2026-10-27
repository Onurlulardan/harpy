package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileCollectionPredicate(
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, error) {
	if len(values) == 0 || len(values) > 100 {
		return "", ErrContract
	}
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", err
	}
	element, cast := api.LogicalTypeText, "pg_catalog.text[]"
	if kind == api.LogicalTypeUUIDList {
		element, cast = api.LogicalTypeUUID, "pg_catalog.uuid[]"
	} else if kind != api.LogicalTypeTextList {
		return "", ErrContract
	}
	placeholders := make([]string, len(values))
	for index, value := range values {
		actual, valueErr := value.Kind()
		if valueErr != nil || actual != element {
			return "", ErrContract
		}
		placeholder, bindErr := bindResolvedValue(value, binder)
		if bindErr != nil {
			return "", bindErr
		}
		placeholders[index] = placeholder
	}
	expression, err := baseExpression(field)
	if err != nil {
		return "", err
	}
	operation := " && "
	if operator == api.OperatorContainsAll {
		operation = " @> "
	} else if operator != api.OperatorContainsAny {
		return "", ErrContract
	}
	array := "CAST(ARRAY[" + strings.Join(placeholders, ", ") + "] AS " + cast + ")"
	return expression + operation + array, nil
}

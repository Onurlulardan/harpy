package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileJSONBExact(
	document string,
	key string,
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, bool, error) {
	if len(values) != 1 || operator != api.OperatorEq && operator != api.OperatorNe {
		return "", true, ErrContract
	}
	bound, err := bindFieldValue(field, values[0], binder)
	if err != nil {
		return "", true, err
	}
	expression := document + " -> " + key
	if operator == api.OperatorEq {
		return expression + " = " + bound, true, nil
	}
	return document + " ? " + key + " AND " + expression + " <> " + bound, true, nil
}

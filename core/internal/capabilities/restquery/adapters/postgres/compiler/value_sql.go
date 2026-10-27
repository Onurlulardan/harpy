package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func bindResolvedValue(value resolved.Value, binder *binder) (string, error) {
	kind, err := value.Kind()
	if err != nil {
		return "", ErrContract
	}
	switch kind {
	case api.LogicalTypeBoolean:
		boolean, present, valueErr := value.Boolean()
		if valueErr != nil || !present {
			return "", ErrContract
		}
		return binder.bind(boolean)
	case api.LogicalTypeJSON:
		document, present, valueErr := value.JSON()
		if valueErr != nil || !present {
			return "", ErrContract
		}
		placeholder, bindErr := binder.bind(string(document))
		if bindErr != nil {
			return "", bindErr
		}
		return "CAST(" + placeholder + " AS pg_catalog.jsonb)", nil
	default:
		scalar, present, valueErr := value.Scalar()
		if valueErr != nil || !present {
			return "", ErrContract
		}
		return binder.bind(scalar)
	}
}

func bindFieldValue(
	field api.FieldDefinition, value resolved.Value, binder *binder,
) (string, error) {
	fieldType, err := field.Type()
	if err != nil {
		return "", ErrContract
	}
	expected, expectedErr := fieldType.Kind()
	actual, actualErr := value.Kind()
	if expectedErr != nil || actualErr != nil || expected != actual {
		return "", ErrContract
	}
	return bindResolvedValue(value, binder)
}

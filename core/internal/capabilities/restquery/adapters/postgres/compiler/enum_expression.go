package compiler

import (
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type orderedEnumValue struct {
	code    string
	ordinal int
}

func enumOrderExpression(field api.FieldDefinition, binder *binder) (string, error) {
	values, err := orderedEnumValues(field)
	if err != nil {
		return "", err
	}
	codes := make([]string, len(values))
	for index, value := range values {
		codes[index] = value.code
	}
	placeholder, err := binder.bind(codes)
	if err != nil {
		return "", err
	}
	column, err := baseExpression(field)
	if err != nil {
		return "", err
	}
	return "pg_catalog.array_position(CAST(" + placeholder +
		" AS pg_catalog.text[]), " + column + ")", nil
}

func orderedEnumValues(field api.FieldDefinition) ([]orderedEnumValue, error) {
	fieldType, err := field.Type()
	if err != nil {
		return nil, ErrContract
	}
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return nil, ErrContract
	}
	ordered, err := profile.Ordered()
	if err != nil || !ordered {
		return nil, ErrContract
	}
	declared, err := profile.Values()
	if err != nil || len(declared) == 0 {
		return nil, ErrContract
	}
	values := make([]orderedEnumValue, len(declared))
	for index, value := range declared {
		code, codeErr := value.Code()
		ordinal, present, ordinalErr := value.Ordinal()
		if codeErr != nil || ordinalErr != nil || !present {
			return nil, ErrContract
		}
		values[index] = orderedEnumValue{code: code, ordinal: ordinal}
	}
	sort.Slice(values, func(left, right int) bool {
		return values[left].ordinal < values[right].ordinal
	})
	return values, nil
}

func bindEnumOrdinal(
	field api.FieldDefinition, value resolved.Value, binder *binder,
) (string, error) {
	scalar, present, err := value.Scalar()
	if err != nil || !present {
		return "", ErrContract
	}
	values, err := orderedEnumValues(field)
	if err != nil {
		return "", err
	}
	for _, declared := range values {
		if declared.code == scalar {
			return binder.bind(declared.ordinal)
		}
	}
	return "", ErrContract
}

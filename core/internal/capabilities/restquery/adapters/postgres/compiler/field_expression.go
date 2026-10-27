package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func appendBaseColumn(builder *strings.Builder, field api.FieldDefinition) error {
	if builder == nil || field.Validate() != nil {
		return ErrContract
	}
	storage, err := field.Storage()
	if err != nil {
		return ErrContract
	}
	column, err := storage.Column()
	if err != nil || appendIdentifier(builder, column) != nil {
		return ErrContract
	}
	return nil
}

func baseExpression(field api.FieldDefinition) (string, error) {
	if expression, jsonb, err := jsonbFieldExpression(field); err != nil {
		return "", err
	} else if jsonb {
		return expression, nil
	}
	var builder strings.Builder
	if appendBaseColumn(&builder, field) != nil {
		return "", ErrContract
	}
	return builder.String(), nil
}

func exactExpression(field api.FieldDefinition) (string, error) {
	expression, err := baseExpression(field)
	if err != nil {
		return "", err
	}
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", err
	}
	if kind == api.LogicalTypeText {
		expression += " COLLATE pg_catalog.pg_unicode_fast"
	}
	return expression, nil
}

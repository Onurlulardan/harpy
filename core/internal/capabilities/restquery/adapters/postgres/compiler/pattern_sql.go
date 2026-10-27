package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compilePattern(
	field api.FieldDefinition,
	operator api.Operator,
	value resolved.Value,
	binder *binder,
) (string, error) {
	expression, err := patternExpression(field)
	if err != nil {
		return "", err
	}
	scalar, present, err := value.Scalar()
	if err != nil || !present || scalar == "" {
		return "", ErrContract
	}
	escaped := escapeLikeLiteral(scalar)
	switch operator {
	case api.OperatorContains:
		escaped = "%" + escaped + "%"
	case api.OperatorStartsWith:
		escaped += "%"
	case api.OperatorEndsWith:
		escaped = "%" + escaped
	default:
		return "", ErrContract
	}
	placeholder, err := binder.bind(escaped)
	if err != nil {
		return "", err
	}
	return expression + " LIKE " + placeholder + " ESCAPE E'\\\\'", nil
}

func patternExpression(field api.FieldDefinition) (string, error) {
	fieldType, typeErr := field.Type()
	profile, found, profileErr := fieldType.Text()
	if typeErr != nil || profileErr != nil || !found {
		return "", ErrContract
	}
	_, casefold, casefoldErr := profile.Casefold()
	_, accent, accentErr := profile.Accent()
	storage, storageErr := field.Storage()
	if casefoldErr != nil || accentErr != nil || storageErr != nil || !casefold {
		return "", ErrContract
	}
	column, present, err := storage.FoldedCompanion()
	if accent {
		column, present, err = storage.AccentCompanion()
	}
	if err != nil || !present {
		return "", ErrContract
	}
	var builder strings.Builder
	if appendIdentifier(&builder, column) != nil {
		return "", ErrContract
	}
	return builder.String() + " COLLATE pg_catalog.pg_unicode_fast", nil
}

func escapeLikeLiteral(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "%", "\\%")
	return strings.ReplaceAll(value, "_", "\\_")
}

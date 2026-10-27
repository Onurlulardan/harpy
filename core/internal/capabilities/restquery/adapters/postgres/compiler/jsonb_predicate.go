package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileJSONBPredicate(
	field api.FieldDefinition,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, bool, error) {
	document, key, present, err := jsonbStorageExpression(field)
	if err != nil || !present {
		return "", present, err
	}
	if operator == api.OperatorIsNull || operator == api.OperatorIsNotNull {
		if len(values) != 0 {
			return "", true, ErrContract
		}
		if operator == api.OperatorIsNull {
			return "NOT (" + document + " ? " + key + ")", true, nil
		}
		return document + " ? " + key, true, nil
	}
	_, kind, identityErr := fieldIdentity(field)
	if identityErr != nil {
		return "", true, identityErr
	}
	if kind == api.LogicalTypeJSON {
		return compileJSONBExact(document, key, field, operator, values, binder)
	}
	if operator == api.OperatorContainsAny || operator == api.OperatorContainsAll {
		return compileJSONBCollection(document, key, operator, values, binder)
	}
	if operator != api.OperatorEq && operator != api.OperatorNe &&
		operator != api.OperatorIn && operator != api.OperatorNotIn {
		return "", true, ErrContract
	}
	terms := make([]string, len(values))
	for index, value := range values {
		bound, bindErr := bindJSONBValue(value, binder)
		if bindErr != nil {
			return "", true, bindErr
		}
		terms[index] = document + " @> pg_catalog.jsonb_build_object(" + key + ", " + bound + ")"
	}
	return combineJSONBTerms(document, key, operator, terms), true, nil
}

func combineJSONBTerms(document, key string, operator api.Operator, terms []string) string {
	joined := "(" + strings.Join(terms, " OR ") + ")"
	if operator == api.OperatorNe || operator == api.OperatorNotIn {
		return document + " ? " + key + " AND NOT " + joined
	}
	return joined
}

func bindJSONBValue(value resolved.Value, binder *binder) (string, error) {
	kind, err := value.Kind()
	if err != nil {
		return "", ErrContract
	}
	bound, err := bindResolvedValue(value, binder)
	if err != nil {
		return "", err
	}
	if kind == api.LogicalTypeBoolean || kind == api.LogicalTypeJSON {
		return bound, nil
	}
	return "CAST(" + bound + " AS pg_catalog.text)", nil
}

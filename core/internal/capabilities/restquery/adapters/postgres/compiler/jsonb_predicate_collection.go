package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileJSONBCollection(
	document string,
	key string,
	operator api.Operator,
	values []resolved.Value,
	binder *binder,
) (string, bool, error) {
	if len(values) == 0 || len(values) > 100 {
		return "", true, ErrContract
	}
	bound := make([]string, len(values))
	for index, value := range values {
		candidate, err := bindJSONBValue(value, binder)
		if err != nil {
			return "", true, err
		}
		bound[index] = candidate
	}
	if operator == api.OperatorContainsAll {
		array := "pg_catalog.jsonb_build_array(" + strings.Join(bound, ", ") + ")"
		return document + " @> pg_catalog.jsonb_build_object(" + key + ", " + array + ")", true, nil
	}
	if operator != api.OperatorContainsAny {
		return "", true, ErrContract
	}
	terms := make([]string, len(bound))
	for index, value := range bound {
		terms[index] = document + " @> pg_catalog.jsonb_build_object(" + key +
			", pg_catalog.jsonb_build_array(" + value + "))"
	}
	return "(" + strings.Join(terms, " OR ") + ")", true, nil
}

package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileSearch(terms []resolved.SearchTerm, binder *binder) (string, error) {
	if len(terms) == 0 || binder == nil {
		return "", ErrContract
	}
	compiled := make([]string, len(terms))
	for index, term := range terms {
		field, fieldErr := term.Field()
		value, valueErr := term.Value()
		if fieldErr != nil || valueErr != nil {
			return "", ErrContract
		}
		predicate, err := compilePattern(field, api.OperatorContains, value, binder)
		if err != nil {
			return "", err
		}
		compiled[index] = parenthesize(predicate)
	}
	return strings.Join(compiled, " OR "), nil
}

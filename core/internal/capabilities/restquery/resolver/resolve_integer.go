package resolver

import (
	"strconv"
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveInteger(
	fieldType api.FieldType, source literalSource,
) (resolved.Value, restqueryv1.IssueCode) {
	text, issue := source.stringValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	if !canonicalIntegerGrammar(text) {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	parsed, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueOutOfRange
	}
	profile, found, err := fieldType.Integer()
	if err != nil || !found {
		return resolved.Value{}, restqueryv1.IssueTypeMismatch
	}
	minimum, minimumErr := profile.Minimum()
	maximum, maximumErr := profile.Maximum()
	if minimumErr != nil || maximumErr != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	if parsed < minimum || parsed > maximum {
		return resolved.Value{}, restqueryv1.IssueValueOutOfRange
	}
	value, err := resolved.NewScalarValue(api.LogicalTypeInteger, text)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

func canonicalIntegerGrammar(value string) bool {
	if value == "" || value == "-0" || value[0] == '+' ||
		len(value) > 1 && value[0] == '0' ||
		len(value) > 1 && value[0] == '-' && value[1] == '0' {
		return false
	}
	digits := strings.TrimPrefix(value, "-")
	if digits == "" {
		return false
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

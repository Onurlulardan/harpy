package resolver

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
)

func resolveText(
	fieldType api.FieldType, operator api.Operator, source literalSource,
) (resolved.Value, restqueryv1.IssueCode) {
	text, issue := source.stringValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	if strings.ContainsRune(text, '\x00') {
		return resolved.Value{}, restqueryv1.IssueNULForbidden
	}
	profile, found, err := fieldType.Text()
	if err != nil || !found {
		return resolved.Value{}, restqueryv1.IssueTypeMismatch
	}
	canonical := textprofile.ExactNFC(text)
	if patternOperator(operator) {
		canonical, issue = normalizePattern(canonical, profile)
		if issue != "" {
			return resolved.Value{}, issue
		}
	}
	value, err := resolved.NewScalarValue(api.LogicalTypeText, canonical)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

func normalizePattern(
	value string, profile api.TextProfile,
) (string, restqueryv1.IssueCode) {
	casefold, enabled, err := profile.Casefold()
	if err != nil || !enabled || casefold != api.TextCasefoldUnicodeDefaultV1 {
		return "", restqueryv1.IssueOperatorNotAllowed
	}
	result := textprofile.UnicodeDefaultCasefoldV1(value)
	accent, hasAccent, err := profile.Accent()
	if err != nil {
		return "", restqueryv1.IssueValueNotCanonical
	}
	if hasAccent {
		if accent != api.TextAccentInsensitiveV1 {
			return "", restqueryv1.IssueValueNotCanonical
		}
		result = textprofile.AccentInsensitiveV1(value)
	}
	if result == "" {
		return "", restqueryv1.IssueValueNotAllowed
	}
	return result, ""
}

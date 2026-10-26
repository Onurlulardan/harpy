package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveEnum(
	fieldType api.FieldType, source literalSource,
) (resolved.Value, restqueryv1.IssueCode) {
	text, issue := source.stringValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return resolved.Value{}, restqueryv1.IssueTypeMismatch
	}
	enumValue, found, err := profile.Lookup(text)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	if !found {
		return resolved.Value{}, restqueryv1.IssueValueNotAllowed
	}
	canonical, err := enumValue.Code()
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	value, err := resolved.NewScalarValue(api.LogicalTypeEnum, canonical)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

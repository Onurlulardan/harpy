package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func resolveUUID(
	fieldKind api.LogicalType, source literalSource,
) (resolved.Value, restqueryv1.IssueCode) {
	text, issue := source.stringValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	parsed, err := uuid.Parse(text)
	if err != nil || parsed == uuid.Nil || parsed.Variant() != uuid.RFC4122 ||
		parsed.String() != text {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	valueKind := fieldKind
	if fieldKind == api.LogicalTypeUUIDList {
		valueKind = api.LogicalTypeUUID
	}
	value, err := resolved.NewScalarValue(valueKind, text)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

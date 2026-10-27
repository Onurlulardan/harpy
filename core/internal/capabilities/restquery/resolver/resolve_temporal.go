package resolver

import (
	"time"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

const timestampLayout = "2006-01-02T15:04:05.000000Z"

func resolveTemporal(
	kind api.LogicalType, source literalSource,
) (resolved.Value, restqueryv1.IssueCode) {
	text, issue := source.stringValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	valid := false
	if kind == api.LogicalTypeDate {
		parsed, err := time.Parse(time.DateOnly, text)
		valid = err == nil && parsed.Format(time.DateOnly) == text
	} else if kind == api.LogicalTypeTimestamp {
		parsed, err := time.Parse(timestampLayout, text)
		valid = err == nil && parsed.UTC().Format(timestampLayout) == text
	}
	if !valid {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	value, err := resolved.NewScalarValue(kind, text)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

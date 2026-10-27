package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveBoolean(source literalSource) (resolved.Value, restqueryv1.IssueCode) {
	boolean, issue := source.booleanValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	value, err := resolved.NewBooleanValue(boolean)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

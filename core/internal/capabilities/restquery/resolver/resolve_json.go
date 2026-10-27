package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveJSON(source literalSource) (resolved.Value, restqueryv1.IssueCode) {
	document := source.jsonDocument()
	hasNUL, valid := jsonDocumentNULState(document)
	if !valid {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	if hasNUL {
		return resolved.Value{}, restqueryv1.IssueNULForbidden
	}
	value, err := resolved.NewJSONValue(document)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

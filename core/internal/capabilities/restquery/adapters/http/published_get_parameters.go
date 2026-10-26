package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

func normalizePublishedGETParameters(
	document publishedGETDocument,
	manifest published.ParameterManifest,
	collector *publishedIssueCollector,
) map[string][]api.Literal {
	result := make(map[string][]api.Literal, len(document.parameters))
	for name, parameter := range document.parameters {
		path := appendJSONPointer("/param", name)
		definition, known, err := manifest.Lookup(name)
		if err != nil || !known {
			collector.add(path, restqueryv1.IssueParameterUnknown)
			continue
		}
		arity, arityErr := definition.Arity()
		if arityErr != nil || parameter.single != nil && parameter.list != nil ||
			arity == published.ParameterAritySingle && parameter.single == nil ||
			arity == published.ParameterArityList && parameter.list == nil {
			collector.add(path, restqueryv1.IssueParameterArityMismatch)
			continue
		}
		if arity == published.ParameterAritySingle {
			literal, literalErr := api.NewTextLiteral(parameter.single.value)
			if literalErr != nil {
				collector.add(path, restqueryv1.IssueValueNotCanonical)
				continue
			}
			result[name] = []api.Literal{literal}
			continue
		}
		result[name] = normalizePublishedGETList(parameter.list, path, collector)
	}
	return result
}

func normalizePublishedGETList(
	values map[int]getMember,
	path string,
	collector *publishedIssueCollector,
) []api.Literal {
	indices, valid := orderedPublishedGETIndices(values, path, collector)
	if len(indices) < 1 || len(indices) > 100 {
		collector.add(path, restqueryv1.IssueParameterArityMismatch)
		valid = false
	}
	result := make([]api.Literal, 0, len(indices))
	for _, index := range indices {
		member := values[index]
		literal, err := api.NewTextLiteral(member.value)
		if err != nil {
			collector.add(member.pointer, restqueryv1.IssueValueNotCanonical)
			valid = false
			continue
		}
		result = append(result, literal)
	}
	if !valid {
		return nil
	}
	return result
}

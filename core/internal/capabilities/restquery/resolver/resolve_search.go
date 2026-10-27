package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveSearch(
	catalog api.FieldCatalog,
	search api.RequestSearch,
	path string,
	collector *resolutionIssues,
) []resolved.SearchTerm {
	fieldIDs, err := catalog.SearchFields()
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return nil
	}
	literal, err := api.NewTextLiteral(search.Term())
	if err != nil {
		collector.add(path, restqueryv1.IssueValueNotCanonical)
		return nil
	}
	terms := make([]resolved.SearchTerm, 0, len(fieldIDs))
	for _, fieldID := range fieldIDs {
		field, found, fieldErr := catalog.Field(fieldID)
		if fieldErr != nil || !found {
			collector.add(path, restqueryv1.IssueReferenceInvalid)
			continue
		}
		fieldType, typeErr := field.Type()
		if typeErr != nil {
			collector.add(path, restqueryv1.IssueTypeMismatch)
			continue
		}
		values, valueErr := ResolveLiterals(
			fieldType, api.OperatorContains, []api.Literal{literal},
		)
		if valueErr != nil || len(values) != 1 {
			collector.add(path, literalIssue(valueErr))
			continue
		}
		term, termErr := resolved.NewSearchTerm(
			field, values[0], mustRequestPath(path),
		)
		if termErr != nil {
			collector.add(path, restqueryv1.IssueCapabilityRequired)
			continue
		}
		terms = append(terms, term)
	}
	return terms
}

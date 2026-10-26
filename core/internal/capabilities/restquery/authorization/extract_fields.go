package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractProjections(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	projections []resolved.Projection,
) error {
	for _, projection := range projections {
		hidden, err := projection.Hidden()
		if err != nil {
			return errPolicyEvaluationContract
		}
		if hidden {
			continue
		}
		field, fieldErr := projection.Field()
		path, pathErr := projection.Path()
		if fieldErr != nil || pathErr != nil || collector.addField(
			RequirementFieldSelect, catalog, resourcePath, path, field,
		) != nil {
			return errPolicyEvaluationContract
		}
	}
	return nil
}

func extractSorts(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	sorts []resolved.Sort,
) error {
	for _, sort := range sorts {
		implicit, err := sort.Implicit()
		if err != nil {
			return errPolicyEvaluationContract
		}
		if implicit {
			continue
		}
		field, fieldErr := sort.Field()
		path, pathErr := sort.Path()
		fieldPath, memberErr := requestMemberPath(path, "/field")
		if fieldErr != nil || pathErr != nil || memberErr != nil ||
			collector.addField(
				RequirementFieldSort, catalog, resourcePath, fieldPath, field,
			) != nil {
			return errPolicyEvaluationContract
		}
	}
	return nil
}

func extractSearch(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	search []resolved.SearchTerm,
) error {
	for _, term := range search {
		field, fieldErr := term.Field()
		path, pathErr := term.Path()
		if fieldErr != nil || pathErr != nil || collector.addField(
			RequirementFieldSearch, catalog, resourcePath, path, field,
		) != nil {
			return errPolicyEvaluationContract
		}
	}
	return nil
}

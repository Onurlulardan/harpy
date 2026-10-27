package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractFilter(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	filter resolved.Filter,
) error {
	if filter.Validate() != nil {
		return errPolicyEvaluationContract
	}
	predicate, found, err := filter.Predicate()
	if err != nil {
		return errPolicyEvaluationContract
	}
	if found {
		return extractFilterPredicate(
			collector, catalog, resourcePath, predicate,
		)
	}
	children, err := filter.Children()
	if err != nil {
		return errPolicyEvaluationContract
	}
	for _, child := range children {
		if err := extractFilter(
			collector, catalog, resourcePath, child,
		); err != nil {
			return err
		}
	}
	return nil
}

func extractFilterPredicate(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	predicate resolved.Predicate,
) error {
	field, fieldErr := predicate.Field()
	path, pathErr := predicate.Path()
	fieldPath, memberErr := requestMemberPath(path, "/field")
	if fieldErr != nil || pathErr != nil || memberErr != nil {
		return errPolicyEvaluationContract
	}
	return collector.addField(
		RequirementFieldFilter, catalog, resourcePath, fieldPath, field,
	)
}

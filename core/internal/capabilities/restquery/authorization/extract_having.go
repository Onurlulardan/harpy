package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractHavingRequirements(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
	having resolved.Having,
) error {
	predicate, found, err := having.Predicate()
	if err != nil {
		return errPolicyEvaluationContract
	}
	if found {
		return extractHavingPredicate(
			collector, catalog, resourcePath, groups, measures, predicate,
		)
	}
	children, err := having.Children()
	if err != nil {
		return errPolicyEvaluationContract
	}
	for _, child := range children {
		if err := extractHavingRequirements(
			collector, catalog, resourcePath, groups, measures, child,
		); err != nil {
			return err
		}
	}
	return nil
}

func extractHavingPredicate(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
	predicate resolved.HavingPredicate,
) error {
	reference, err := predicate.Reference()
	if err != nil {
		return errPolicyEvaluationContract
	}
	field, hasField, err := referencedAggregateField(reference, groups, measures)
	path, pathErr := reference.Path()
	requestPath, memberErr := requestMemberPath(path, "/name")
	if err != nil || pathErr != nil || memberErr != nil {
		return errPolicyEvaluationContract
	}
	if hasField {
		return collector.addField(
			RequirementFieldHaving, catalog, resourcePath, requestPath, field,
		)
	}
	return collector.addAction(
		RequirementAggregateHaving, catalog, resourcePath, requestPath,
	)
}

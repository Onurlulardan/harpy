package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractAggregateRequirements(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	query resolved.AggregateQuery,
) error {
	groups, groupErr := query.Groups()
	measures, measureErr := query.Measures()
	search, searchErr := query.Search()
	if groupErr != nil || measureErr != nil || searchErr != nil ||
		extractAggregateGroups(collector, catalog, resourcePath, groups) != nil ||
		extractAggregateMeasures(collector, catalog, resourcePath, measures) != nil ||
		extractSearch(collector, catalog, resourcePath, search) != nil {
		return errPolicyEvaluationContract
	}
	where, hasWhere, whereErr := query.Where()
	if whereErr != nil || hasWhere && extractFilter(
		collector, catalog, resourcePath, where,
	) != nil {
		return errPolicyEvaluationContract
	}
	having, hasHaving, havingErr := query.Having()
	if havingErr != nil || hasHaving && extractHavingRequirements(
		collector, catalog, resourcePath, groups, measures, having,
	) != nil {
		return errPolicyEvaluationContract
	}
	sorts, err := query.Sorts()
	if err != nil {
		return errPolicyEvaluationContract
	}
	return extractAggregateSorts(
		collector, catalog, resourcePath, groups, measures, sorts,
	)
}

func extractAggregateGroups(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	groups []resolved.AggregateGroup,
) error {
	for _, group := range groups {
		field, fieldErr := group.Field()
		path, pathErr := group.Path()
		if fieldErr != nil || pathErr != nil || collector.addField(
			RequirementFieldGroup, catalog, resourcePath, path, field,
		) != nil {
			return errPolicyEvaluationContract
		}
	}
	return nil
}

func extractAggregateMeasures(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	measures []resolved.AggregateMeasure,
) error {
	for _, measure := range measures {
		field, hasField, fieldErr := measure.Field()
		path, pathErr := measure.Path()
		if fieldErr != nil || pathErr != nil {
			return errPolicyEvaluationContract
		}
		member := "/function"
		if hasField {
			member = "/field"
		}
		requestPath, err := requestMemberPath(path, member)
		if err != nil {
			return err
		}
		if hasField {
			if collector.addField(
				RequirementFieldMeasure, catalog, resourcePath, requestPath, field,
			) != nil {
				return errPolicyEvaluationContract
			}
		} else if collector.addAction(
			RequirementAggregateMeasure, catalog, resourcePath, requestPath,
		) != nil {
			return errPolicyEvaluationContract
		}
	}
	return nil
}

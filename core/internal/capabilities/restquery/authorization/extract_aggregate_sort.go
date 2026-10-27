package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func extractAggregateSorts(
	collector *requirementCollector,
	catalog api.FieldCatalog,
	resourcePath resolved.RequestPath,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
	sorts []resolved.AggregateSort,
) error {
	for _, sort := range sorts {
		implicit, err := sort.Implicit()
		if err != nil {
			return errPolicyEvaluationContract
		}
		if implicit {
			continue
		}
		reference, err := sort.Reference()
		if err != nil {
			return errPolicyEvaluationContract
		}
		field, hasField, fieldErr := referencedAggregateField(
			reference, groups, measures,
		)
		path, pathErr := reference.Path()
		requestPath, memberErr := requestMemberPath(path, "/name")
		if fieldErr != nil || pathErr != nil || memberErr != nil {
			return errPolicyEvaluationContract
		}
		if hasField {
			if collector.addField(
				RequirementFieldSort, catalog, resourcePath, requestPath, field,
			) != nil {
				return errPolicyEvaluationContract
			}
		} else if collector.addAction(
			RequirementAggregateSort, catalog, resourcePath, requestPath,
		) != nil {
			return errPolicyEvaluationContract
		}
	}
	return nil
}

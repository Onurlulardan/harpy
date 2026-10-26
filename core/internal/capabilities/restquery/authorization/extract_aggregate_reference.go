package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func referencedAggregateField(
	reference resolved.AggregateReference,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
) (api.FieldDefinition, bool, error) {
	kind, kindErr := reference.Kind()
	ordinal, ordinalErr := reference.Ordinal()
	if kindErr != nil || ordinalErr != nil || ordinal < 0 {
		return api.FieldDefinition{}, false, errPolicyEvaluationContract
	}
	switch kind {
	case api.AggregateReferenceGroup:
		if ordinal >= len(groups) {
			return api.FieldDefinition{}, false, errPolicyEvaluationContract
		}
		field, err := groups[ordinal].Field()
		return field, err == nil, err
	case api.AggregateReferenceMeasure:
		if ordinal >= len(measures) {
			return api.FieldDefinition{}, false, errPolicyEvaluationContract
		}
		return measures[ordinal].Field()
	default:
		return api.FieldDefinition{}, false, errPolicyEvaluationContract
	}
}

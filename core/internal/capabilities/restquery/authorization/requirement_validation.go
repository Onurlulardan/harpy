package authorization

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func validRequirementTarget(requirement Requirement) bool {
	field := requirement.fieldID != uuid.Nil && requirement.fieldID.Variant() == uuid.RFC4122
	relation := requirement.relationID != uuid.Nil &&
		requirement.relationID.Variant() == uuid.RFC4122
	switch requirement.kind {
	case RequirementFieldSelect, RequirementFieldFilter, RequirementFieldSort,
		RequirementFieldSearch, RequirementFieldGroup, RequirementFieldMeasure,
		RequirementFieldHaving:
		return field && !relation
	case RequirementRelationInclude:
		return !field && relation
	case RequirementResourceRead, RequirementExactCount, RequirementAggregateMeasure,
		RequirementAggregateHaving, RequirementAggregateSort:
		return !field && !relation
	default:
		return false
	}
}

func requestPathBelongsToResource(
	resourcePath resolved.RequestPath,
	requestPath resolved.RequestPath,
) bool {
	resource, resourceErr := resourcePath.Value()
	request, requestErr := requestPath.Value()
	if resourceErr != nil || requestErr != nil || resource == "" {
		return resourceErr == nil && requestErr == nil
	}
	return request == resource || strings.HasPrefix(request, resource+"/")
}

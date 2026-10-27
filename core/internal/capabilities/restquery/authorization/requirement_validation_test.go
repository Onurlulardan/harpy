package authorization

import (
	"testing"

	"github.com/google/uuid"
)

func TestRequirementTargetShapesAreExact(t *testing.T) {
	fieldID := uuid.MustParse(testFieldID)
	relationID := uuid.MustParse(testRelationID)
	fieldKinds := []RequirementKind{
		RequirementFieldSelect, RequirementFieldFilter, RequirementFieldSort,
		RequirementFieldSearch, RequirementFieldGroup, RequirementFieldMeasure,
		RequirementFieldHaving,
	}
	for _, kind := range fieldKinds {
		mustRequirement(t, kind, "", "/select/0", fieldID, uuid.Nil)
		if _, err := newRequirement(
			kind, requirementCatalog(t), requirementPath(t, ""),
			requirementPath(t, "/select/0"), uuid.Nil, relationID,
		); err == nil {
			t.Fatalf("%s accepted relation target", kind)
		}
	}

	mustRequirement(
		t, RequirementRelationInclude, "", "/include/0", uuid.Nil, relationID,
	)
	resourceKinds := []RequirementKind{
		RequirementResourceRead, RequirementExactCount, RequirementAggregateMeasure,
		RequirementAggregateHaving, RequirementAggregateSort,
	}
	for _, kind := range resourceKinds {
		mustRequirement(t, kind, "", "/aggregate", uuid.Nil, uuid.Nil)
	}
}

func TestRequirementRequestPathMustBelongToResourceNode(t *testing.T) {
	_, err := newRequirement(
		RequirementResourceRead,
		requirementCatalog(t),
		requirementPath(t, "/include/0"),
		requirementPath(t, "/include/1/select/0"),
		uuid.Nil,
		uuid.Nil,
	)
	if err == nil {
		t.Fatal("cross-node request path was accepted")
	}
}

package authorization

import (
	"fmt"
	"slices"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestExtractRecordRequirementsCoversEveryActorGatedUse(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	set, err := extractRequirements(authorizationRecordSpec(t, fixture))
	if err != nil || !set.validate() {
		t.Fatalf("extract record requirements: %v", err)
	}
	want := []string{
		"resource_read|forms|||--|--",
		"field_select|forms||/select/0|" + authRootAmount + "|--",
		"field_filter|forms||/where/field|" + authRootAmount + "|--",
		"field_sort|forms||/sort/0/field|" + authRootName + "|--",
		"field_search|forms||/search|" + authRootName + "|--",
		"exact_count|forms||/count|--|--",
		"relation_include|forms||/include/0/relation|--|" + authChildren,
		"resource_read|crm|/include/0|/include/0|--|--",
		"field_select|crm|/include/0|/include/0/select/0|" + authTargetName + "|--",
		"field_filter|crm|/include/0|/include/0/where/field|" + authTargetName + "|--",
		"field_sort|crm|/include/0|/include/0/sort/0/field|" + authTargetName + "|--",
	}
	requireExactRequirementLabels(t, set, want)
}

func TestExtractAggregateRequirementsSeparatesFieldlessActions(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	set, err := extractRequirements(authorizationAggregateSpec(t, fixture))
	if err != nil || !set.validate() {
		t.Fatalf("extract aggregate requirements: %v", err)
	}
	want := []string{
		"resource_read|forms|||--|--",
		"field_group|forms||/aggregate/group_by/0|" + authRootName + "|--",
		"field_measure|forms||/aggregate/measures/0/field|" + authRootAmount + "|--",
		"aggregate_measure|forms||/aggregate/measures/1/function|--|--",
		"field_filter|forms||/where/field|" + authRootAmount + "|--",
		"field_search|forms||/search|" + authRootName + "|--",
		"field_having|forms||/aggregate/having/and/0/name|" + authRootName + "|--",
		"aggregate_having|forms||/aggregate/having/and/1/name|--|--",
		"field_sort|forms||/aggregate/sort/0/name|" + authRootName + "|--",
		"aggregate_sort|forms||/aggregate/sort/1/name|--|--",
	}
	requireExactRequirementLabels(t, set, want)
}

func TestExtractRequirementsRejectsInvalidSpec(t *testing.T) {
	if _, err := extractRequirements(resolved.QuerySpec{}); err == nil {
		t.Fatal("invalid QuerySpec was accepted")
	}
}

func requireExactRequirementLabels(
	t *testing.T,
	set requirementSet,
	want []string,
) {
	t.Helper()
	actual := make([]string, 0, len(set.requirements))
	for _, requirement := range set.requirements {
		actual = append(actual, requirementLabel(t, requirement))
	}
	slices.Sort(actual)
	slices.Sort(want)
	if !slices.Equal(actual, want) {
		t.Fatalf("requirements\nactual=%q\nwant=%q", actual, want)
	}
}

func requirementLabel(t *testing.T, requirement Requirement) string {
	t.Helper()
	kind, _ := requirement.Kind()
	catalog, _ := requirement.Catalog()
	owner, _ := catalog.Owner()
	resourcePath, _ := requirement.ResourcePath()
	requestPath, _ := requirement.RequestPath()
	resource, _ := resourcePath.Value()
	request, _ := requestPath.Value()
	field, hasField, _ := requirement.FieldID()
	relation, hasRelation, _ := requirement.RelationID()
	fieldValue, relationValue := "--", "--"
	if hasField {
		fieldValue = field.String()
	}
	if hasRelation {
		relationValue = relation.String()
	}
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s",
		kind, owner, resource, request, fieldValue, relationValue,
	)
}

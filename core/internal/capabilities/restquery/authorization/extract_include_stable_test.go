package authorization

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func TestExtractIncludeStableRequirementMatchesVisibility(t *testing.T) {
	tests := []struct {
		name         string
		selectFields []string
		wantStable   bool
	}{
		{name: "automatic stable is internal", selectFields: []string{"name"}},
		{name: "explicit stable is gated", selectFields: []string{"id"}, wantStable: true},
		{name: "default stable is gated", selectFields: nil, wantStable: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set, err := extractRequirements(includeStableSpec(t, test.selectFields))
			if err != nil {
				t.Fatal(err)
			}
			present := hasStableSelectRequirement(t, set)
			if present != test.wantStable {
				t.Fatalf("stable requirement present = %v", present)
			}
		})
	}
}

func hasStableSelectRequirement(t *testing.T, set requirementSet) bool {
	t.Helper()
	for _, requirement := range set.requirements {
		kind, _ := requirement.Kind()
		field, hasField, _ := requirement.FieldID()
		if kind == RequirementFieldSelect && hasField && field.String() == authTargetID {
			return true
		}
	}
	return false
}

func includeStableSpec(t *testing.T, selectFields []string) resolved.QuerySpec {
	t.Helper()
	fixture := newAuthorizationQueryFixture(t)
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: selectFields,
	})
	if err != nil {
		t.Fatal(err)
	}
	page, _ := api.NewRequestPage(10, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"amount"}, Include: []api.RequestInclude{include},
		Page: page, Count: api.CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, fixture.catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

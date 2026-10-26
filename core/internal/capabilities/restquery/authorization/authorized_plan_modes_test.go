package authorization

import (
	"bytes"
	"testing"
)

func TestBuildAuthorizedPlanPreservesAllAndNoneVisibility(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	withNone := make([]policyEvaluation, 0, len(evaluations))
	for _, evaluation := range evaluations {
		request := evaluation.request
		withNone = append(withNone, evaluationWithVisibility(
			t, request, func(requirement Requirement) Visibility {
				if request.owner == "forms" {
					visibility, _ := NewNoVisibility(requirement.id)
					return visibility
				}
				visibility, _ := NewAllVisibility(requirement.id)
				return visibility
			},
		))
	}
	plan, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, withNone,
	)
	if err != nil {
		t.Fatal(err)
	}
	visibilities, err := plan.Visibilities()
	if err != nil || len(visibilities) != 2 {
		t.Fatalf("plan visibilities: %v", err)
	}
	modes := make(map[string]string, len(visibilities))
	for _, visibility := range visibilities {
		catalog, catalogErr := visibility.Catalog()
		owner, ownerErr := catalog.Owner()
		mode, modeErr := visibility.Mode()
		_, hasFilter, filterErr := visibility.Filter()
		if catalogErr != nil || ownerErr != nil || modeErr != nil ||
			filterErr != nil || hasFilter {
			t.Fatal("all/none visibility gained an invalid predicate")
		}
		modes[owner] = mode.String()
	}
	if modes["forms"] != "none" || modes["crm"] != "all" {
		t.Fatalf("visibility modes = %#v", modes)
	}
	canonical, err := plan.MandatoryCanonical()
	if err != nil || bytes.Contains(canonical, []byte(`"filter"`)) {
		t.Fatal("all/none canonical document contained a filter")
	}
}

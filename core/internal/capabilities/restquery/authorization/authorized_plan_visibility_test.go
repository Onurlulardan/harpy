package authorization

import (
	"bytes"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestMandatoryVisibilityCanonicalizesCommutativePolicyOrder(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	amount := mandatoryPredicateFilter(
		t, authRootAmount, api.LogicalTypeDecimal, api.OperatorGT,
		resolvedScalar(t, api.LogicalTypeDecimal, "1.20"),
	)
	name := mandatoryPredicateFilter(
		t, authRootName, api.LogicalTypeText, api.OperatorEq,
		resolvedScalar(t, api.LogicalTypeText, "Ada"),
	)
	forward, _ := NewMandatoryAnd([]MandatoryFilter{amount, name})
	reverse, _ := NewMandatoryAnd([]MandatoryFilter{name, amount})
	first := evaluationsWithRootFilter(t, evaluations, forward)
	second := evaluationsWithRootFilter(t, evaluations, reverse)
	firstPlan, firstErr := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, first,
	)
	secondPlan, secondErr := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, second,
	)
	firstCanonical, _ := firstPlan.MandatoryCanonical()
	secondCanonical, _ := secondPlan.MandatoryCanonical()
	if firstErr != nil || secondErr != nil ||
		!bytes.Equal(firstCanonical, secondCanonical) {
		t.Fatal("commutative mandatory policy order changed canonical identity")
	}
	if !bytes.Contains(firstCanonical, []byte(`"value":"1.20"`)) {
		t.Fatal("mandatory canonical document lost typed predicate")
	}
}

func TestMandatoryVisibilityRejectsCatalogProfileAndFieldDrift(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	cases := []MandatoryFilter{
		mandatoryPredicateFilter(
			t, authRootAmount, api.LogicalTypeDecimal, api.OperatorGT,
			resolvedScalar(t, api.LogicalTypeDecimal, "99999999999.99"),
		),
		mandatoryPredicateFilter(
			t, authTargetName, api.LogicalTypeText, api.OperatorEq,
			resolvedScalar(t, api.LogicalTypeText, "Ada"),
		),
	}
	for index, filter := range cases {
		withFilter := evaluationsWithRootFilter(t, evaluations, filter)
		if _, err := buildAuthorizedPlan(
			spec, set, authorizedPlanTenant, authorizedPlanActor, withFilter,
		); err == nil {
			t.Fatalf("invalid mandatory filter %d was authorized", index)
		}
	}
}

func evaluationsWithRootFilter(
	t *testing.T,
	evaluations []policyEvaluation,
	filter MandatoryFilter,
) []policyEvaluation {
	t.Helper()
	result := make([]policyEvaluation, 0, len(evaluations))
	for _, evaluation := range evaluations {
		request := evaluation.request
		result = append(result, evaluationWithVisibility(
			t, request, func(requirement Requirement) Visibility {
				if request.owner != "forms" {
					visibility, _ := NewAllVisibility(requirement.id)
					return visibility
				}
				visibility, err := NewPredicateVisibility(requirement.id, filter)
				if err != nil {
					t.Fatal(err)
				}
				return visibility
			},
		))
	}
	return result
}

func mandatoryPredicateFilter(
	t *testing.T,
	fieldID string,
	kind api.LogicalType,
	operator api.Operator,
	value resolved.Value,
) MandatoryFilter {
	t.Helper()
	predicate, err := NewMandatoryPredicate(
		fieldID, kind, operator, []resolved.Value{value},
	)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := NewMandatoryPredicateFilter(predicate)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func resolvedScalar(
	t *testing.T, kind api.LogicalType, value string,
) resolved.Value {
	t.Helper()
	result, err := resolved.NewScalarValue(kind, value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

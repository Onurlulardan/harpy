package authorization

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestBuildAuthorizedPlanRejectsCrossBoundInputs(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	partial := removeOneNonResourceRequirement(t, set)
	wrongQuery := authorizationEvaluations(
		t, set, authorizedPlanTenant, authorizedPlanActor,
		sha256.Sum256([]byte("different-query")),
	)
	otherTenant := uuid.MustParse("71000000-0000-4000-8000-000000000002")
	otherActor := uuid.MustParse("72000000-0000-4000-8000-000000000002")
	tests := []struct {
		name        string
		set         requirementSet
		tenant      uuid.UUID
		actor       uuid.UUID
		evaluations []policyEvaluation
	}{
		{"partial requirement set", partial, authorizedPlanTenant, authorizedPlanActor, evaluations},
		{"different query decision", set, authorizedPlanTenant, authorizedPlanActor, wrongQuery},
		{"different tenant", set, otherTenant, authorizedPlanActor, evaluations},
		{"different actor", set, authorizedPlanTenant, otherActor, evaluations},
		{"zero tenant", set, uuid.Nil, authorizedPlanActor, evaluations},
		{"zero actor", set, authorizedPlanTenant, uuid.Nil, evaluations},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := buildAuthorizedPlan(
				spec, test.set, test.tenant, test.actor, test.evaluations,
			); err == nil {
				t.Fatal("cross-bound authorization inputs minted a plan")
			}
		})
	}
}

func TestBuildAuthorizedPlanRejectsTransplantedDecision(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	otherTenant := uuid.MustParse("71000000-0000-4000-8000-000000000002")
	identity, err := spec.IdentityHash()
	if err != nil {
		t.Fatal(err)
	}
	foreign := authorizationEvaluations(
		t, set, otherTenant, authorizedPlanActor, identity,
	)
	transplanted := append([]policyEvaluation{}, evaluations...)
	transplanted[0].decision = foreign[0].decision
	if _, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, transplanted,
	); err == nil {
		t.Fatal("decision from another tenant binding minted a plan")
	}
}

func removeOneNonResourceRequirement(
	t *testing.T, set requirementSet,
) requirementSet {
	t.Helper()
	owned := make([]Requirement, 0, len(set.requirements)-1)
	removed := false
	for _, requirement := range set.requirements {
		if !removed && requirement.kind != RequirementResourceRead {
			removed = true
			continue
		}
		owned = append(owned, requirement)
	}
	set.requirements = owned
	if !removed || !set.validate() {
		t.Fatal("fixture did not produce a valid partial requirement set")
	}
	return set
}

func authorizationEvaluations(
	t *testing.T,
	set requirementSet,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	queryIdentity [sha256.Size]byte,
) []policyEvaluation {
	t.Helper()
	requests, err := groupPolicyRequests(
		tenantID, actorID, queryIdentity, set.requirements,
	)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]policyEvaluation, 0, len(requests))
	for _, request := range requests {
		decision := exactDecision(t, request, nil)
		validated, err := validatePolicyDecision(request, decision)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, policyEvaluation{
			owner: request.owner, request: request, decision: validated,
		})
	}
	return result
}

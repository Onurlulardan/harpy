package authorization

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestBuildAuthorizedPlanReturnsExactFieldDenial(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	denied := denyFirstRequirement(t, evaluations, "forms", RequirementFieldSelect)
	_, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, denied,
	)
	var queryError api.QueryError
	if !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorQueryFieldForbidden {
		t.Fatalf("field denial = %v", err)
	}
	issues := queryError.Issues()
	if len(issues) != 1 || issues[0].Path != "/select/0" ||
		issues[0].Code != string(restqueryv1.IssueFieldForbidden) {
		t.Fatalf("field denial issues = %#v", issues)
	}
}

func TestBuildAuthorizedPlanHidesResourceDenialDetails(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	denied := denyFirstRequirement(t, evaluations, "", RequirementResourceRead)
	_, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, denied,
	)
	var queryError api.QueryError
	if !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorQueryForbidden ||
		len(queryError.Issues()) != 0 {
		t.Fatalf("resource denial leaked: %v %#v", err, queryError.Issues())
	}
}

func denyFirstRequirement(
	t *testing.T,
	evaluations []policyEvaluation,
	owner string,
	kind RequirementKind,
) []policyEvaluation {
	t.Helper()
	result := append([]policyEvaluation{}, evaluations...)
	for index, evaluation := range result {
		if owner != "" && evaluation.request.owner != owner {
			continue
		}
		for _, requirement := range evaluation.request.requirements {
			if requirement.kind != kind {
				continue
			}
			decision := exactDecision(
				t, evaluation.request, map[[32]byte]bool{requirement.id: true},
			)
			validated, err := validatePolicyDecision(evaluation.request, decision)
			if err != nil {
				t.Fatal(err)
			}
			result[index].decision = validated
			return result
		}
	}
	t.Fatalf("requirement kind %s not found", kind)
	return nil
}

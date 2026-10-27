package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func buildAuthorizedPlan(
	spec resolved.QuerySpec,
	set requirementSet,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	evaluations []policyEvaluation,
) (resolved.AuthorizedPlan, error) {
	validatedSet, err := validateAuthorizedPlanInputs(
		spec, set, tenantID, actorID, evaluations,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	denied := make([]Requirement, 0)
	for _, evaluation := range evaluations {
		if !evaluation.validate() {
			return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
		}
		denied = append(denied, evaluation.decision.denied...)
	}
	queryError, found, err := queryErrorForDenials(denied)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	if found {
		return resolved.AuthorizedPlan{}, queryError
	}
	authorizationContext, err := authorizationContextSHA256(evaluations)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	visibilities, err := resolvePolicyVisibilities(validatedSet, evaluations)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, tenantID, actorID, authorizationContext, visibilities,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	return plan, nil
}

package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func buildGlobalAuthorizedPlan(
	spec resolved.QuerySpec,
	set requirementSet,
	request GlobalPolicyRequest,
	decision validatedDecision,
) (resolved.AuthorizedPlan, error) {
	if spec.Validate() != nil || !set.validate() || request.Validate() != nil ||
		!sameDigest(request.binding, decision.requestBinding) {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	queryError, found, err := queryErrorForDenials(decision.denied)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	if found {
		return resolved.AuthorizedPlan{}, queryError
	}
	contextHash, err := globalAuthorizationContextSHA256(request, decision)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	visibilities, err := resolveGlobalVisibilities(set, decision)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	plan, err := resolved.MintGlobalAuthorizedPlan(
		spec, request.environmentID, request.actorID, request.operationID,
		request.public, request.snapshotAt, contextHash, visibilities,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, errPolicyEvaluationContract
	}
	return plan, nil
}

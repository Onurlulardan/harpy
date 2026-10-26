package authorization

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type policyEvaluation struct {
	owner    string
	request  PolicyRequest
	decision validatedDecision
}

func evaluatePolicyRequests(
	ctx context.Context,
	scope transaction.TenantScope,
	registry PolicyRegistry,
	requests []PolicyRequest,
) ([]policyEvaluation, error) {
	if ctx == nil || transaction.RequireActive(scope, transaction.SnapshotQuery) != nil ||
		transaction.NetworkAllowed(ctx) || ctx.Err() != nil {
		return nil, errPolicyEvaluationContract
	}
	ordered, err := orderedPolicyRequests(
		registry, scope.TenantID(), scope.ActorID(), requests,
	)
	if err != nil {
		return nil, err
	}
	evaluations := make([]policyEvaluation, 0, len(ordered))
	for _, request := range ordered {
		port, found := registry.lookup(request.owner)
		if !found {
			return nil, errPolicyEvaluationContract
		}
		decision, callErr := port.EvaluateQuery(ctx, scope, request)
		if callErr != nil {
			return nil, newPolicyEvaluationFailure(callErr)
		}
		if ctx.Err() != nil ||
			transaction.RequireActive(scope, transaction.SnapshotQuery) != nil {
			return nil, errPolicyEvaluationContract
		}
		validated, validationErr := validatePolicyDecision(request, decision)
		if validationErr != nil {
			return nil, errPolicyDecisionContract
		}
		evaluations = append(evaluations, policyEvaluation{
			owner: request.owner, request: request, decision: validated,
		})
	}
	return evaluations, nil
}

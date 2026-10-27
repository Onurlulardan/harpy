package reports

import (
	"context"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Policy struct{ catalogs *Catalog }

func NewPolicy(catalogs *Catalog) (*Policy, error) {
	if catalogs.Validate() != nil {
		return nil, ErrContract
	}
	return &Policy{catalogs: catalogs}, nil
}

func (policy *Policy) EvaluateGlobalQuery(
	ctx context.Context,
	scope transaction.GlobalScope,
	request restauthorization.GlobalPolicyRequest,
) (restauthorization.PolicyDecision, error) {
	if policy == nil || policy.catalogs.Validate() != nil || ctx == nil ||
		ctx.Err() != nil || request.Validate() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireGlobalOwner(scope, Owner) != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	authority, err := transaction.MintGlobalSnapshotAuthority(scope, Owner)
	if err != nil || !authority.Owns(scope) {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	requirements, err := request.Requirements()
	if err != nil || len(requirements) == 0 {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	spec, results, resourceID, err := policyResults(policy.catalogs, requirements)
	if err != nil || policyAdmission(scope, request, spec) != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	allowed := allowedReportOperation(spec, scope.OperationID())
	for index := range results {
		identity, identityErr := requirements[index].ID()
		if identityErr != nil {
			return restauthorization.PolicyDecision{}, ErrContract
		}
		results[index], err = restauthorization.NewDecisionResult(identity, allowed)
		if err != nil {
			return restauthorization.PolicyDecision{}, ErrContract
		}
	}
	visibilities := []restauthorization.Visibility{}
	if allowed {
		visibility, visibilityErr := restauthorization.NewAllVisibility(resourceID)
		if visibilityErr != nil {
			return restauthorization.PolicyDecision{}, ErrContract
		}
		visibilities = append(visibilities, visibility)
	}
	binding, err := request.BindingSHA256()
	revision, revisionErr := policyRevision(
		policy.catalogs, spec, authority, scope.OperationID(),
	)
	if err != nil || revisionErr != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	decision, err := restauthorization.NewPolicyDecision(
		binding, revision, results, visibilities,
	)
	if err != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	return decision, nil
}

func allowedReportOperation(spec resourceSpec, operation string) bool {
	if operation == spec.Operation {
		return true
	}
	return spec.Name == SubscriptionResource &&
		operation == SubscriptionTransitionOperation
}

package offering

import (
	"context"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const platformListOperation = "billing.catalog.offering.list"

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
		ctx.Err() != nil || request.Validate() != nil ||
		transaction.NetworkAllowed(ctx) || !transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireGlobalOwner(scope, Owner) != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	authority, err := transaction.MintGlobalSnapshotAuthority(scope, Owner)
	if err != nil || !authority.Owns(scope) {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	public, operation, allowed, err := policyAdmission(scope, request)
	if err != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	requirements, err := request.Requirements()
	if err != nil || len(requirements) == 0 {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	results, resourceID, err := policyResults(policy.catalogs, requirements, allowed)
	if err != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	visibilities, err := policyVisibilities(resourceID, public, allowed, request)
	if err != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	binding, err := request.BindingSHA256()
	revision, revisionErr := policyRevision(policy.catalogs, authority, public, operation)
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

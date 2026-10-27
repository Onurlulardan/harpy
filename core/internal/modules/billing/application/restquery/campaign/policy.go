package campaign

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
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
	if err != nil || !authority.Owns(scope) || !admitted(scope, request) {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	requirements, err := request.Requirements()
	if err != nil || len(requirements) == 0 {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	results, resourceID, err := policyResults(policy.catalogs, requirements)
	if err != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	visibility, err := restauthorization.NewAllVisibility(resourceID)
	binding, bindingErr := request.BindingSHA256()
	revision, revisionErr := policyRevision(policy.catalogs, authority)
	if err != nil || bindingErr != nil || revisionErr != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	decision, err := restauthorization.NewPolicyDecision(
		binding, revision, results, []restauthorization.Visibility{visibility},
	)
	if err != nil {
		return restauthorization.PolicyDecision{}, ErrContract
	}
	return decision, nil
}

func admitted(scope transaction.GlobalScope, request restauthorization.GlobalPolicyRequest) bool {
	environmentID, environmentErr := request.EnvironmentID()
	actorID, hasActor, actorErr := request.ActorID()
	operation, operationErr := request.OperationID()
	public, publicErr := request.Public()
	owner, ownerErr := request.Owner()
	return environmentErr == nil && actorErr == nil && operationErr == nil &&
		publicErr == nil && ownerErr == nil && environmentID == scope.EnvironmentID() &&
		owner == Owner && !public && hasActor && actorID != uuid.Nil &&
		actorID == scope.ActorID() && operation == platformListOperation
}

func policyResults(
	catalog *Catalog, requirements []restauthorization.Requirement,
) ([]restauthorization.DecisionResult, [sha256.Size]byte, error) {
	results := make([]restauthorization.DecisionResult, len(requirements))
	var resourceID [sha256.Size]byte
	for index, requirement := range requirements {
		reference, referenceErr := requirement.Catalog()
		kind, kindErr := requirement.Kind()
		identity, identityErr := requirement.ID()
		if referenceErr != nil || kindErr != nil || identityErr != nil ||
			!catalog.Matches(reference) || !allowedRequirement(kind) {
			return nil, [sha256.Size]byte{}, ErrContract
		}
		if kind == restauthorization.RequirementResourceRead {
			if resourceID != [sha256.Size]byte{} {
				return nil, [sha256.Size]byte{}, ErrContract
			}
			resourceID = identity
		}
		result, err := restauthorization.NewDecisionResult(identity, true)
		if err != nil {
			return nil, [sha256.Size]byte{}, ErrContract
		}
		results[index] = result
	}
	if resourceID == [sha256.Size]byte{} {
		return nil, [sha256.Size]byte{}, ErrContract
	}
	return results, resourceID, nil
}

func allowedRequirement(kind restauthorization.RequirementKind) bool {
	switch kind {
	case restauthorization.RequirementResourceRead,
		restauthorization.RequirementFieldSelect,
		restauthorization.RequirementFieldFilter,
		restauthorization.RequirementFieldSort,
		restauthorization.RequirementFieldSearch,
		restauthorization.RequirementExactCount:
		return true
	default:
		return false
	}
}

func policyRevision(
	catalog *Catalog, authority transaction.GlobalSnapshotAuthority,
) ([sha256.Size]byte, error) {
	reference, err := catalog.Reference()
	if err != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	hash, hashErr := reference.CatalogHash()
	hashHex, hexErr := hash.Hex()
	document := struct {
		Version, EnvironmentID, CatalogSHA256, Operation string
		GroupGeneration, BillingGeneration               int64
	}{
		"billing-campaign-rest-query-policy.v1", authority.EnvironmentID().String(),
		hashHex, platformListOperation, authority.GroupGeneration(),
		authority.BillingGeneration(),
	}
	canonical, marshalErr := json.Marshal(document)
	if hashErr != nil || hexErr != nil || marshalErr != nil ||
		document.GroupGeneration < 1 || document.BillingGeneration < 1 {
		return [sha256.Size]byte{}, ErrContract
	}
	return sha256.Sum256(canonical), nil
}

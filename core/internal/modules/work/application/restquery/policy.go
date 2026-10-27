package restquery

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"sort"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type workQueryPolicy struct {
	catalogs   map[string]ResolvedCatalog
	principals workprincipalapi.Directory
}

func (policy *workQueryPolicy) EvaluateQuery(ctx context.Context,
	scope platformtx.TenantScope, request restauthorization.PolicyRequest,
) (restauthorization.PolicyDecision, error) {
	evidence, ok := evidenceFrom(ctx)
	if !ok || !workQueryScopeValid(ctx, scope, request, evidence) {
		return restauthorization.PolicyDecision{}, ErrHTTPUnavailable
	}
	requirements, err := request.Requirements()
	if err != nil || len(requirements) == 0 {
		return restauthorization.PolicyDecision{}, ErrHTTPUnavailable
	}
	catalog, resourceID, err := policy.matchRequirements(requirements, evidence.resource)
	if err != nil {
		return restauthorization.PolicyDecision{}, err
	}
	current, err := policy.principals.Evidence(ctx, scope, "USER",
		[16]byte(scope.ActorID()))
	if err != nil || current.AuthorizationRevision < 1 {
		return restauthorization.PolicyDecision{}, ErrHTTPUnavailable
	}
	keys, keysOK := workPrincipalKeys(current.Principals)
	allowed := current.TargetExists && keysOK &&
		current.AuthorizationRevision == evidence.authorizationVersion
	results, err := workQueryResults(requirements, allowed)
	if err != nil {
		return restauthorization.PolicyDecision{}, err
	}
	visibility, err := workQueryVisibility(resourceID, catalog, evidence,
		keys, current.AuthorizationRevision, allowed)
	if err != nil {
		return restauthorization.PolicyDecision{}, err
	}
	binding, err := request.BindingSHA256()
	revision, revisionErr := workQueryRevision(catalog, evidence, current, keys, scope)
	if err != nil || revisionErr != nil {
		return restauthorization.PolicyDecision{}, ErrHTTPUnavailable
	}
	return restauthorization.NewPolicyDecision(binding, revision, results,
		[]restauthorization.Visibility{visibility})
}

func workQueryScopeValid(ctx context.Context, scope platformtx.TenantScope,
	request restauthorization.PolicyRequest, evidence Evidence,
) bool {
	if ctx == nil || ctx.Err() != nil || request.Validate() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		platformtx.RequireOwner(scope, Owner) != nil || !evidence.valid() ||
		scope.OperationID() != evidence.operation {
		return false
	}
	tenant, tenantErr := request.TenantID()
	actor, actorErr := request.ActorID()
	owner, ownerErr := request.Owner()
	return tenantErr == nil && actorErr == nil && ownerErr == nil && owner == Owner &&
		tenant == scope.TenantID() && tenant == evidence.tenantID &&
		actor == scope.ActorID() && actor == evidence.actorID
}

type workQueryRevisionDocument struct {
	Contract, Actor, Tenant, Resource, Catalog string
	Token, Authorization, Generation           int64
	Principals                                 []string
}

func workQueryRevision(catalog ResolvedCatalog, evidence Evidence,
	current workprincipalapi.Evidence, keys []string, scope platformtx.TenantScope,
) ([sha256.Size]byte, error) {
	authority, err := platformtx.MintTenantSnapshotAuthority(scope, "work")
	hash, hashErr := catalog.reference.CatalogHash()
	hashHex, hexErr := hash.Hex()
	document := workQueryRevisionDocument{"work-rest-query-policy.v1",
		evidence.actorID.String(), evidence.tenantID.String(),
		evidence.resource, hashHex, evidence.tokenVersion,
		current.AuthorizationRevision, authority.Generation(), append([]string{}, keys...)}
	canonical, marshalErr := json.Marshal(document)
	if err != nil || hashErr != nil || hexErr != nil || marshalErr != nil ||
		authority.Generation() < 1 {
		return [sha256.Size]byte{}, ErrHTTPUnavailable
	}
	return sha256.Sum256(canonical), nil
}

func workPrincipalKeys(principals []workprincipalapi.Principal) ([]string, bool) {
	if len(principals) == 0 || len(principals) > 100 {
		return nil, false
	}
	keys := make([]string, 0, len(principals))
	seen := make(map[string]struct{}, len(principals))
	for _, principal := range principals {
		if principal.ID == [16]byte{} || principal.Revision < 1 || (principal.Kind != "USER" &&
			principal.Kind != "ROLE" && principal.Kind != "ORGANIZATION") {
			return nil, false
		}
		key := principal.Kind + ":" + uuid.UUID(principal.ID).String()
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key], keys = struct{}{}, append(keys, key)
	}
	sort.Strings(keys)
	return keys, len(keys) > 0
}

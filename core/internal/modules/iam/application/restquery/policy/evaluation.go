package policy

import (
	"context"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	iamauthorization "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type permissionKey struct {
	resource string
	action   string
}

type evaluationState struct {
	service     *Service
	ctx         context.Context
	scope       transaction.TenantScope
	versions    ActorVersions
	catalogs    *catalogCache
	permissions map[permissionKey]iamauthorization.Decision
	clock       int64
	clockSet    bool
}

func newEvaluationState(
	service *Service,
	ctx context.Context,
	scope transaction.TenantScope,
	versions ActorVersions,
) *evaluationState {
	return &evaluationState{
		service: service, ctx: ctx, scope: scope, versions: versions,
		catalogs:    newCatalogCache(service, ctx, scope),
		permissions: make(map[permissionKey]iamauthorization.Decision),
	}
}

func (state *evaluationState) evaluate(
	requirement restauthorization.Requirement,
) (restauthorization.DecisionResult, *restauthorization.Visibility, error) {
	reference, err := requirement.Catalog()
	if err != nil {
		return restauthorization.DecisionResult{}, nil, catalog.ErrContract
	}
	published, err := state.catalogs.load(reference)
	if err != nil {
		return restauthorization.DecisionResult{}, nil, catalog.ErrContract
	}
	resource, err := published.Resource()
	action, actionErr := requirementAction(requirement, published)
	if err != nil || actionErr != nil {
		return restauthorization.DecisionResult{}, nil, catalog.ErrContract
	}
	decision, err := state.permission(resource.Key(), action)
	if err != nil {
		return restauthorization.DecisionResult{}, nil, err
	}
	requirementID, err := requirement.ID()
	if err != nil {
		return restauthorization.DecisionResult{}, nil, catalog.ErrContract
	}
	result, err := restauthorization.NewDecisionResult(requirementID, decision.Allowed())
	if err != nil {
		return restauthorization.DecisionResult{}, nil, catalog.ErrContract
	}
	kind, _ := requirement.Kind()
	if kind != restauthorization.RequirementResourceRead {
		return result, nil, nil
	}
	visibility, err := resourceVisibility(requirementID, resource, decision.Allowed())
	if err != nil {
		return restauthorization.DecisionResult{}, nil, err
	}
	return result, &visibility, nil
}

func (state *evaluationState) permission(
	resourceKey string,
	action string,
) (iamauthorization.Decision, error) {
	key := permissionKey{resource: resourceKey, action: action}
	if decision, found := state.permissions[key]; found {
		return decision, nil
	}
	request, err := iamauthorization.NewRequest(
		resourceKey, action, state.versions.TokenVersion(),
		state.versions.AuthorizationVersion(),
	)
	if err != nil {
		return iamauthorization.Decision{}, catalog.ErrContract
	}
	decision, err := state.service.authorizer.Authorize(state.ctx, state.scope, request)
	if err != nil || decision.Validate() != nil ||
		decision.AuthorizationVersion() != state.versions.AuthorizationVersion() {
		return iamauthorization.Decision{}, catalog.ErrContract
	}
	if state.clockSet && state.clock != decision.AuthorizationClock() {
		return iamauthorization.Decision{}, catalog.ErrContract
	}
	state.clock, state.clockSet = decision.AuthorizationClock(), true
	state.permissions[key] = decision
	return decision, nil
}

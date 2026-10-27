package policy

import (
	"context"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Authorize(
	ctx context.Context,
	scope transaction.TenantScope,
	request restauthorization.PolicyRequest,
) (restauthorization.PolicyDecision, error) {
	if service == nil || interfaceNil(service.authorizer) ||
		interfaceNil(service.catalogs) || interfaceNil(service.versions) ||
		!requestScopeUsable(ctx, scope, request) {
		return restauthorization.PolicyDecision{}, catalog.ErrContract
	}
	actorID, _ := request.ActorID()
	versions, err := service.versions.Versions(ctx, scope, actorID)
	if err != nil || versions.Validate() != nil || versions.ActorID() != actorID {
		return restauthorization.PolicyDecision{}, catalog.ErrContract
	}
	requirements, err := request.Requirements()
	if err != nil || len(requirements) == 0 {
		return restauthorization.PolicyDecision{}, catalog.ErrContract
	}
	state := newEvaluationState(service, ctx, scope, versions)
	results := make([]restauthorization.DecisionResult, 0, len(requirements))
	visibilities := make([]restauthorization.Visibility, 0, 3)
	for _, requirement := range requirements {
		result, visibility, evaluationErr := state.evaluate(requirement)
		if evaluationErr != nil {
			return restauthorization.PolicyDecision{}, evaluationErr
		}
		results = append(results, result)
		if visibility != nil {
			visibilities = append(visibilities, *visibility)
		}
	}
	revision, err := contextRevision(state)
	binding, bindingErr := request.BindingSHA256()
	if err != nil || bindingErr != nil {
		return restauthorization.PolicyDecision{}, catalog.ErrContract
	}
	decision, err := restauthorization.NewPolicyDecision(
		binding, revision, results, visibilities,
	)
	if err != nil {
		return restauthorization.PolicyDecision{}, catalog.ErrContract
	}
	return decision, nil
}

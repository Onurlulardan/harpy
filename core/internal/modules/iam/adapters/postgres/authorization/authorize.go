package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
)

func (authorizer *authorizer) Authorize(
	ctx context.Context,
	scope tenantScope,
	request authorization.Request,
) (authorization.Snapshot, error) {
	if authorizer == nil || !authorizer.valid || request.Validate() != nil ||
		!authorizationScopeUsable(ctx, scope, authorizer.mode) {
		return authorization.Snapshot{}, errAuthorizationRead
	}
	bound, err := authorizer.bind(scope)
	if err != nil {
		return authorization.Snapshot{}, errAuthorizationRead
	}
	actorID := scope.ActorID()
	if explicit, ok := request.ActorID(); ok {
		actorID = explicit
	}
	return bound.readEvidence(ctx, scope.TenantID(), actorID, request)
}

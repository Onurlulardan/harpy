package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *ActorBridgeIssuerIAM) AuthorizeActorBridge(
	ctx context.Context,
	scope platformtx.TenantScope,
	resourceKey string,
	actionKey string,
	tokenVersion int64,
	authorizationVersion int64,
) (platformactorbridge.Authorization, error) {
	if !runtime.Valid() {
		return platformactorbridge.Authorization{},
			platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	request, err := authorization.NewRequest(
		resourceKey, actionKey, tokenVersion, authorizationVersion,
	)
	if err != nil {
		return platformactorbridge.Authorization{},
			platformactorbridge.NewFault(platformactorbridge.FaultInvalid)
	}
	decision, err := runtime.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return platformactorbridge.Authorization{},
			platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	if !decision.Allowed() {
		return platformactorbridge.Authorization{},
			platformactorbridge.NewFault(platformactorbridge.FaultUnauthorized)
	}
	permissionID, allowed := decision.PermissionID()
	if !allowed || decision.AuthorizationVersion() != authorizationVersion {
		return platformactorbridge.Authorization{},
			platformactorbridge.NewFault(platformactorbridge.FaultUnauthorized)
	}
	result, err := platformactorbridge.NewAuthorization(
		scope.TenantID(), scope.ActorID(), permissionID, tokenVersion,
		decision.AuthorizationVersion(), decision.AuthorizationClock(),
	)
	if err != nil {
		return platformactorbridge.Authorization{},
			platformactorbridge.NewFault(platformactorbridge.FaultUnavailable)
	}
	return result, nil
}

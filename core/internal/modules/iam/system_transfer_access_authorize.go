package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferAccessRuntime) authorizeSystemTransfer(ctx context.Context,
	claims accesstoken.Claims, resource SystemTransferResource,
	action string) (authorization.Decision, error) {
	request, err := authorization.NewRequest(
		"iam."+string(resource), action,
		claims.TokenVersion(), claims.AuthorizationVersion())
	if err != nil {
		return authorization.Decision{}, ErrSystemTransferForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrSystemTransferUnavailable
	}
	var result authorization.Decision
	err = runtime.coordinator.InTenant(ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.authorizer.Authorize(scoped, scope, request)
			if authorizeErr != nil || decision.Validate() != nil {
				return ErrSystemTransferUnavailable
			}
			if !decision.Allowed() {
				return ErrSystemTransferForbidden
			}
			if _, found := decision.PermissionID(); !found ||
				decision.AuthorizationClock() < 1 {
				return ErrSystemTransferUnavailable
			}
			result = decision
			return nil
		})
	if errors.Is(err, ErrSystemTransferForbidden) {
		return authorization.Decision{}, ErrSystemTransferForbidden
	}
	if err != nil {
		return authorization.Decision{}, ErrSystemTransferUnavailable
	}
	return result, nil
}

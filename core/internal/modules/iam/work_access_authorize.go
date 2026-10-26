package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *WorkAccessRuntime) authorize(
	ctx context.Context, claims accesstoken.Claims, target workAccessTarget,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(
		target.resource, target.action, claims.TokenVersion(), claims.AuthorizationVersion(),
	)
	if err != nil {
		return authorization.Decision{}, ErrWorkAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.shared.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrWorkAccessUnavailable
	}
	var result authorization.Decision
	err = runtime.shared.coordinator.InTenant(ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.shared.authorizer.Authorize(
				scopeContext, scope, request,
			)
			if authorizeErr != nil || decision.Validate() != nil {
				return ErrWorkAccessUnavailable
			}
			if !decision.Allowed() {
				return ErrWorkAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrWorkAccessUnavailable
			}
			result = decision
			return nil
		})
	if err != nil {
		if errors.Is(err, ErrWorkAccessForbidden) {
			return authorization.Decision{}, ErrWorkAccessForbidden
		}
		return authorization.Decision{}, ErrWorkAccessUnavailable
	}
	return result, nil
}

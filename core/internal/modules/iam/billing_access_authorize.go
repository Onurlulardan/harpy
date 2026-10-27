package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *BillingAccessRuntime) authorizeBilling(
	ctx context.Context,
	claims accesstoken.Claims,
	target billingPermissionTarget,
) (authorization.Decision, error) {
	if target.global {
		return authorization.Decision{}, ErrBillingAccessForbidden
	}
	request, err := authorization.NewRequest(
		target.resource, target.action,
		claims.TokenVersion(), claims.AuthorizationVersion(),
	)
	if err != nil {
		return authorization.Decision{}, ErrBillingAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrBillingAccessUnavailable
	}
	var result authorization.Decision
	err = runtime.coordinator.InTenant(
		ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			if runtime.epoch == nil ||
				runtime.epoch.Assert(scopeContext, scope, claims.SecurityEpoch()) != nil {
				return ErrBillingAccessForbidden
			}
			decision, authorizeErr := runtime.authorizer.Authorize(
				scopeContext, scope, request,
			)
			if authorizeErr != nil || decision.Validate() != nil {
				return ErrBillingAccessUnavailable
			}
			if !decision.Allowed() {
				return ErrBillingAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrBillingAccessUnavailable
			}
			result = decision
			return nil
		},
	)
	if err != nil {
		if errors.Is(err, ErrBillingAccessForbidden) {
			return authorization.Decision{}, ErrBillingAccessForbidden
		}
		return authorization.Decision{}, ErrBillingAccessUnavailable
	}
	return result, nil
}

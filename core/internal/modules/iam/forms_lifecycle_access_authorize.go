package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *FormsLifecycleAccessRuntime) authorize(
	ctx context.Context, claims accesstoken.Claims, target formsLifecycleAccessTarget,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(target.resource, target.action,
		claims.TokenVersion(), claims.AuthorizationVersion())
	if err != nil {
		return authorization.Decision{}, ErrFormsLifecycleAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.shared.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrFormsLifecycleAccessUnavailable
	}
	var result authorization.Decision
	err = runtime.shared.coordinator.InTenant(ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.shared.authorizer.Authorize(scopeContext, scope, request)
			if authorizeErr != nil {
				return fmt.Errorf("%w: %w", ErrFormsLifecycleAccessUnavailable, authorizeErr)
			}
			if validateErr := decision.Validate(); validateErr != nil {
				return fmt.Errorf("%w: %w", ErrFormsLifecycleAccessUnavailable, validateErr)
			}
			if !decision.Allowed() {
				return ErrFormsLifecycleAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrFormsLifecycleAccessUnavailable
			}
			result = decision
			return nil
		})
	if err != nil {
		return authorization.Decision{}, mapAccessAuthorizeError(
			ErrFormsLifecycleAccessForbidden, ErrFormsLifecycleAccessUnavailable, err)
	}
	return result, nil
}

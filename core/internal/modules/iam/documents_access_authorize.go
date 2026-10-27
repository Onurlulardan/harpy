package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *DocumentsAccessRuntime) authorize(ctx context.Context,
	claims accesstoken.Claims, target documentsAccessTarget,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(target.resource, target.action,
		claims.TokenVersion(), claims.AuthorizationVersion())
	if err != nil {
		return authorization.Decision{}, ErrDocumentsAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.shared.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrDocumentsAccessUnavailable
	}
	var result authorization.Decision
	err = runtime.shared.coordinator.InTenant(ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.shared.authorizer.Authorize(scopeContext, scope, request)
			if authorizeErr != nil {
				return fmt.Errorf("%w: %w", ErrDocumentsAccessUnavailable, authorizeErr)
			}
			if validateErr := decision.Validate(); validateErr != nil {
				return fmt.Errorf("%w: %w", ErrDocumentsAccessUnavailable, validateErr)
			}
			if !decision.Allowed() {
				return ErrDocumentsAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrDocumentsAccessUnavailable
			}
			result = decision
			return nil
		})
	if err != nil {
		return authorization.Decision{}, mapAccessAuthorizeError(
			ErrDocumentsAccessForbidden, ErrDocumentsAccessUnavailable, err)
	}
	return result, nil
}

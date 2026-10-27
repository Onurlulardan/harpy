package iam

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *WorkflowAccessRuntime) authorize(
	ctx context.Context, claims accesstoken.Claims, target workflowAccessTarget,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(target.resource, target.action,
		claims.TokenVersion(), claims.AuthorizationVersion())
	if err != nil {
		return authorization.Decision{}, ErrWorkflowAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.shared.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrWorkflowAccessUnavailable
	}
	var result authorization.Decision
	err = runtime.shared.coordinator.InTenant(ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.shared.authorizer.Authorize(scopeContext, scope, request)
			if authorizeErr != nil {
				return fmt.Errorf("%w: %w", ErrWorkflowAccessUnavailable, authorizeErr)
			}
			if validateErr := decision.Validate(); validateErr != nil {
				return fmt.Errorf("%w: %w", ErrWorkflowAccessUnavailable, validateErr)
			}
			if !decision.Allowed() {
				return ErrWorkflowAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrWorkflowAccessUnavailable
			}
			result = decision
			return nil
		})
	if err != nil {
		return authorization.Decision{}, mapAccessAuthorizeError(
			ErrWorkflowAccessForbidden, ErrWorkflowAccessUnavailable, err)
	}
	return result, nil
}

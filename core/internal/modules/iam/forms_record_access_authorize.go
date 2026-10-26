package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *FormsRecordAccessRuntime) authorizeFormsRecord(
	ctx context.Context,
	claims accesstoken.Claims,
	target formsRecordAccessTarget,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(
		"forms.records", target.action,
		claims.TokenVersion(), claims.AuthorizationVersion(),
	)
	if err != nil {
		return authorization.Decision{}, ErrFormsRecordAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.pool, claims)
	if err != nil {
		return authorization.Decision{}, ErrFormsRecordAccessUnavailable
	}
	var result authorization.Decision
	err = runtime.coordinator.InTenant(
		ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.authorizer.Authorize(
				scopeContext, scope, request,
			)
			if authorizeErr != nil || decision.Validate() != nil {
				return ErrFormsRecordAccessUnavailable
			}
			if !decision.Allowed() {
				return ErrFormsRecordAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrFormsRecordAccessUnavailable
			}
			result = decision
			return nil
		},
	)
	if err != nil {
		if errors.Is(err, ErrFormsRecordAccessForbidden) {
			return authorization.Decision{}, ErrFormsRecordAccessForbidden
		}
		return authorization.Decision{}, ErrFormsRecordAccessUnavailable
	}
	return result, nil
}

package iam

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemLayoutAccessRuntime) authorizeSystemLayout(
	ctx context.Context,
	claims accesstoken.Claims,
	resource publicresource.Resource,
	target systemLayoutAccessTarget,
) error {
	resourceKey, err := resource.AuthorityResource()
	if err != nil {
		return ErrSystemLayoutAccessForbidden
	}
	request, err := authorization.NewRequest(
		resourceKey, target.permissionAction,
		claims.TokenVersion(), claims.AuthorizationVersion(),
	)
	if err != nil {
		return ErrSystemLayoutAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.pool, claims)
	if err != nil {
		return ErrSystemLayoutAccessUnavailable
	}
	err = runtime.coordinator.InTenant(
		ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.authorizer.Authorize(scoped, scope, request)
			if authorizeErr != nil || decision.Validate() != nil {
				return ErrSystemLayoutAccessUnavailable
			}
			if !decision.Allowed() {
				return ErrSystemLayoutAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrSystemLayoutAccessUnavailable
			}
			return nil
		},
	)
	if errors.Is(err, ErrSystemLayoutAccessForbidden) {
		return ErrSystemLayoutAccessForbidden
	}
	if err != nil {
		return ErrSystemLayoutAccessUnavailable
	}
	return nil
}

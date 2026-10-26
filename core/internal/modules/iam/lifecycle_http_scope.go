package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) commandScope(
	ctx context.Context, claims accesstoken.Claims, operation string,
) (platformtx.ScopeRequest, error) {
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || len(generations) != 2 || generations["iam"] <= 0 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, ErrLifecycleHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(claims.Subject(), claims.Tenant(),
		operation, identityAuthorityGroup, generations["iam"], generations)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrLifecycleHTTPUnavailable
	}
	return request, nil
}

func (application *lifecycleHTTPApplication) snapshotScope(
	ctx context.Context, claims accesstoken.Claims,
) (platformtx.ScopeRequest, error) {
	request, err := iamHTTPSnapshotScopeRequest(ctx, application.pool, claims)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrLifecycleHTTPUnavailable
	}
	return request, nil
}

func (application *lifecycleHTTPApplication) authorize(
	ctx context.Context, scope platformtx.TenantScope,
	claims accesstoken.Claims, action string,
) (uuid.UUID, error) {
	request, err := authorization.NewRequest("iam.users", action,
		claims.TokenVersion(), claims.AuthorizationVersion())
	if err != nil {
		return uuid.Nil, ErrLifecycleHTTPForbidden
	}
	decision, err := application.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return uuid.Nil, ErrLifecycleHTTPUnavailable
	}
	permissionID, allowed := decision.PermissionID()
	if !allowed {
		return uuid.Nil, ErrLifecycleHTTPForbidden
	}
	return permissionID, nil
}

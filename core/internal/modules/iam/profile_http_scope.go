package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *profileHTTPApplication) commandScopeRequest(
	ctx context.Context,
	claims accesstoken.Claims,
	operation string,
) (platformtx.ScopeRequest, error) {
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || len(generations) != 2 || generations["iam"] <= 0 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, ErrProfileHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(
		claims.Subject(), claims.Tenant(), operation, identityAuthorityGroup,
		generations["iam"], generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrProfileHTTPUnavailable
	}
	return request, nil
}

func (application *profileHTTPApplication) snapshotScopeRequest(
	ctx context.Context,
	claims accesstoken.Claims,
) (platformtx.ScopeRequest, error) {
	request, err := iamHTTPSnapshotScopeRequest(ctx, application.pool, claims)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrProfileHTTPUnavailable
	}
	return request, nil
}

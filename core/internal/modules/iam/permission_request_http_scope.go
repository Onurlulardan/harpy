package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *permissionRequestHTTPApplication) commandScopeRequest(
	ctx context.Context,
	claims accesstoken.Claims,
	operationID string,
) (platformtx.ScopeRequest, error) {
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || len(generations) != 2 || generations["iam"] <= 0 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, ErrPermissionRequestHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(
		claims.Subject(), claims.Tenant(), operationID,
		identityAuthorityGroup, generations["iam"], generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrPermissionRequestHTTPUnavailable
	}
	return request, nil
}

func (application *permissionRequestHTTPApplication) snapshotScopeRequest(
	ctx context.Context,
	evidence identityEvidenceSource,
) (platformtx.ScopeRequest, error) {
	request, err := iamHTTPSnapshotScopeRequest(ctx, application.pool, evidence)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrPermissionRequestHTTPUnavailable
	}
	return request, nil
}

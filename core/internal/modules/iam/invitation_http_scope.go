package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *invitationHTTPApplication) commandScope(
	ctx context.Context,
	claims accesstoken.Claims,
	operation string,
) (platformtx.ScopeRequest, error) {
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || len(generations) != 2 || generations["iam"] <= 0 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, ErrInvitationHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(
		claims.Subject(), claims.Tenant(), operation, identityAuthorityGroup,
		generations["iam"], generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrInvitationHTTPUnavailable
	}
	return request, nil
}

func (application *invitationHTTPApplication) snapshotScope(
	ctx context.Context,
	claims accesstoken.Claims,
) (platformtx.ScopeRequest, error) {
	request, err := iamHTTPSnapshotScopeRequest(ctx, application.pool, claims)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrInvitationHTTPUnavailable
	}
	return request, nil
}

func validInvitationClaims(claims accesstoken.Claims) bool {
	return claims.Subject() != [16]byte{} && claims.Tenant() != [16]byte{} &&
		claims.TokenVersion() > 0 && claims.AuthorizationVersion() > 0
}

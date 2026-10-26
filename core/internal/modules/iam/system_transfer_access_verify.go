package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *SystemTransferAccessRuntime) VerifySystemTransfer(ctx context.Context,
	request SystemTransferAccessRequest) (SystemTransferAccess, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil {
		return SystemTransferAccess{}, ErrSystemTransferUnavailable
	}
	target, targetOK := resolveSystemTransferAccess(
		request.Resource, request.Permission, request.Operation)
	wire, bearerOK := formsRecordBearer(request.AuthorizationValues)
	now := runtime.clock().UTC().Truncate(time.Second)
	if !targetOK || !bearerOK || now.IsZero() || now.Unix() <= 0 {
		return SystemTransferAccess{}, ErrSystemTransferAuthentication
	}
	claims, err := runtime.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) || !tenantAccessHeadersMatch(
		ctx, runtime.pool, request.TenantIDValues,
		request.TenantSlugValues, claims.Tenant()) {
		return SystemTransferAccess{}, ErrSystemTransferAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return SystemTransferAccess{}, ErrSystemTransferForbidden
	}
	decision, err := runtime.authorizeSystemTransfer(ctx, claims,
		request.Resource, target.action)
	if err != nil {
		return SystemTransferAccess{}, err
	}
	return NewSystemTransferAccess(claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(),
		decision.AuthorizationClock(), request.Resource,
		request.Permission, request.Operation)
}

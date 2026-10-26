package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
)

func (runtime *WorkAccessRuntime) Verify(
	ctx context.Context, request WorkAccessRequest,
) (WorkAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || ctx == nil || ctx.Err() != nil {
		return WorkAccessDecision{}, ErrWorkAccessUnavailable
	}
	target, targetValid := resolveWorkAccess(request.Authority, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return WorkAccessDecision{}, ErrWorkAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!tenantAccessHeadersMatch(ctx, runtime.shared.pool, request.TenantIDValues,
			request.TenantSlugValues, claims.Tenant()) {
		return WorkAccessDecision{}, ErrWorkAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return WorkAccessDecision{}, ErrWorkAccessForbidden
	}
	authorizationDecision, err := runtime.authorize(ctx, claims, target)
	if err != nil {
		return WorkAccessDecision{}, err
	}
	decision, err := NewProtectedWorkAccessDecision(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), authorizationDecision.AuthorizationClock(),
		authorizationDecision.Reason() == authorization.ReasonProtectedTenantAll,
	)
	if err != nil {
		return WorkAccessDecision{}, ErrWorkAccessUnavailable
	}
	return decision, nil
}

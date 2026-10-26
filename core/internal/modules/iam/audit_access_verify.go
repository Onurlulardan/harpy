package iam

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *AuditAccessRuntime) Verify(
	ctx context.Context, request AuditAccessRequest,
) (AuditAccessDecision, error) {
	if runtime == nil || runtime.base == nil || ctx == nil || ctx.Err() != nil {
		return AuditAccessDecision{}, ErrAuditAccessUnavailable
	}
	target, targetValid := resolveAuditAccess(request.Permission, request.Operation)
	wire, bearerValid := billingBearer(request.AuthorizationValues)
	now := runtime.base.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return AuditAccessDecision{}, ErrAuditAccessAuthentication
	}
	claims, err := runtime.base.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!tenantAccessHeadersMatch(ctx, runtime.base.pool,
			request.TenantIDValues, request.TenantSlugValues, claims.Tenant()) {
		return AuditAccessDecision{}, ErrAuditAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return AuditAccessDecision{}, ErrAuditAccessForbidden
	}
	authority, err := runtime.base.authorizeBilling(ctx, claims, billingPermissionTarget{
		resource: target.resource, action: target.action,
	})
	if err != nil {
		if errors.Is(err, ErrBillingAccessForbidden) {
			return AuditAccessDecision{}, ErrAuditAccessForbidden
		}
		return AuditAccessDecision{}, ErrAuditAccessUnavailable
	}
	decision, err := NewAuditAccessDecision(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), authority.AuthorizationClock(),
		request.Permission, request.Operation,
	)
	if err != nil {
		return AuditAccessDecision{}, ErrAuditAccessUnavailable
	}
	return decision, nil
}

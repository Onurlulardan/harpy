package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *NotificationsAccessRuntime) Verify(ctx context.Context,
	request NotificationsAccessRequest,
) (NotificationsAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || runtime.sessions == nil ||
		ctx == nil || ctx.Err() != nil {
		return NotificationsAccessDecision{}, ErrNotificationsAccessUnavailable
	}
	target, targetValid := resolveNotificationsAccess(request)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	origin, originValid := notificationsOrigin(runtime, request.OriginValues, target.origin)
	if !targetValid || !bearerValid || !originValid || now.IsZero() || now.Unix() <= 0 {
		return NotificationsAccessDecision{}, ErrNotificationsAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!tenantAccessHeadersMatch(ctx, runtime.shared.pool, request.TenantIDValues,
			request.TenantSlugValues, claims.Tenant()) {
		return NotificationsAccessDecision{}, ErrNotificationsAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return NotificationsAccessDecision{}, ErrNotificationsAccessForbidden
	}
	_, selector, err := runtime.authorize(ctx, claims, target, now)
	if err != nil {
		return NotificationsAccessDecision{}, err
	}
	defer clear(selector[:])
	csrfVerified := false
	if target.csrf {
		cookie, header, valid := notificationsCSRF(request.CSRFValues, request.CookieValues)
		if !valid {
			return NotificationsAccessDecision{}, ErrNotificationsAccessAuthentication
		}
		defer clear(cookie)
		defer clear(header)
		if runtime.shared.security.csrfVerifier.Verify(ctx, cookie, header,
			selector, origin, now) != nil {
			return NotificationsAccessDecision{}, ErrNotificationsAccessAuthentication
		}
		csrfVerified = true
	}
	decision, err := NewNotificationsAccessDecision(claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(), origin,
		target.session, csrfVerified)
	if err != nil {
		return NotificationsAccessDecision{}, ErrNotificationsAccessUnavailable
	}
	return decision, nil
}

package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *AnalyticsAccessRuntime) Verify(
	ctx context.Context, request AnalyticsAccessRequest,
) (AnalyticsAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || ctx == nil || ctx.Err() != nil {
		return AnalyticsAccessDecision{}, ErrAnalyticsAccessUnavailable
	}
	target, targetValid := resolveAnalyticsAccess(request.Permission, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return AnalyticsAccessDecision{}, ErrAnalyticsAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return AnalyticsAccessDecision{}, ErrAnalyticsAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return AnalyticsAccessDecision{}, ErrAnalyticsAccessForbidden
	}
	stepUpAt, stepUpValid := analyticsStepUpTime(claims, request.Operation, now)
	if !stepUpValid {
		return AnalyticsAccessDecision{}, ErrAnalyticsAccessForbidden
	}
	authorizationDecision, err := runtime.authorize(ctx, claims, target)
	if err != nil {
		return AnalyticsAccessDecision{}, err
	}
	decision, err := NewAnalyticsAccessDecision(claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(),
		authorizationDecision.AuthorizationClock(), stepUpAt)
	if err != nil {
		return AnalyticsAccessDecision{}, ErrAnalyticsAccessUnavailable
	}
	return decision, nil
}

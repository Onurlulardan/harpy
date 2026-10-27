package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *BillingAccessRuntime) Verify(
	ctx context.Context,
	request BillingAccessRequest,
) (BillingAccessDecision, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil {
		return BillingAccessDecision{}, ErrBillingAccessUnavailable
	}
	target, valid := resolveBillingPermission(request.Permission, request.Operation)
	wire, bearerValid := billingBearer(request.AuthorizationValues)
	now := runtime.clock().UTC().Truncate(time.Second)
	if !valid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return BillingAccessDecision{}, ErrBillingAccessAuthentication
	}
	claims, err := runtime.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return BillingAccessDecision{}, ErrBillingAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return BillingAccessDecision{}, ErrBillingAccessForbidden
	}
	if target.global {
		platformDecision, authorizeErr := runtime.authorizePlatformBilling(
			ctx, claims, target,
		)
		if authorizeErr != nil {
			return BillingAccessDecision{}, authorizeErr
		}
		decision, decisionErr := runtime.globalBillingDecision(
			ctx, claims, request, target, platformDecision,
		)
		if decisionErr != nil {
			return BillingAccessDecision{}, decisionErr
		}
		return decision, nil
	}
	authorizationDecision, err := runtime.authorizeBilling(ctx, claims, target)
	if err != nil {
		return BillingAccessDecision{}, err
	}
	decision, err := NewTenantBillingAccessDecision(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), authorizationDecision.AuthorizationClock(),
	)
	if err != nil {
		return BillingAccessDecision{}, ErrBillingAccessUnavailable
	}
	return decision, nil
}

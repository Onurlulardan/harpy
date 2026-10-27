package iam

import (
	"context"
	"errors"
	"time"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func (runtime *UsageAccessRuntime) Verify(
	ctx context.Context, request UsageAccessRequest,
) (UsageAccessDecision, error) {
	if runtime == nil || runtime.base == nil || ctx == nil || ctx.Err() != nil {
		return UsageAccessDecision{}, ErrUsageAccessUnavailable
	}
	target, targetValid := resolveUsageAccess(
		request.Scope, request.Permission, request.Operation,
	)
	wire, bearerValid := billingBearer(request.AuthorizationValues)
	now := runtime.base.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return UsageAccessDecision{}, ErrUsageAccessAuthentication
	}
	claims, err := runtime.base.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!tenantAccessHeadersMatch(ctx, runtime.base.pool,
			request.TenantIDValues, request.TenantSlugValues, claims.Tenant()) {
		return UsageAccessDecision{}, ErrUsageAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return UsageAccessDecision{}, ErrUsageAccessForbidden
	}
	if target.platform {
		return runtime.verifyPlatform(ctx, claims, request, target)
	}
	return runtime.verifyTenant(ctx, claims, request, target)
}

func (runtime *UsageAccessRuntime) verifyTenant(
	ctx context.Context, claims accesstoken.Claims,
	request UsageAccessRequest, target usageAccessTarget,
) (UsageAccessDecision, error) {
	authority, err := runtime.base.authorizeBilling(ctx, claims, billingPermissionTarget{
		resource: target.resource, action: target.action,
	})
	if err != nil {
		return UsageAccessDecision{}, mapUsageBillingError(err)
	}
	session, err := runtime.base.readConsumerSession(ctx, claims)
	if err != nil {
		if errors.Is(err, errConsumerSessionRejected) {
			return UsageAccessDecision{}, ErrUsageAccessForbidden
		}
		return UsageAccessDecision{}, ErrUsageAccessUnavailable
	}
	decision, err := NewUsageAccessDecision(
		claims.Subject(), claims.Tenant(), claims.Tenant(), runtime.environmentID,
		claims.TokenVersion(), claims.AuthorizationVersion(), authority.AuthorizationClock(),
		request.Scope, string(session.DeploymentMode), string(session.TenantPurpose), false,
		request.Permission, request.Operation,
	)
	if err != nil {
		return UsageAccessDecision{}, ErrUsageAccessUnavailable
	}
	return decision, nil
}

func (runtime *UsageAccessRuntime) verifyPlatform(
	ctx context.Context, claims accesstoken.Claims,
	request UsageAccessRequest, target usageAccessTarget,
) (UsageAccessDecision, error) {
	authority, err := runtime.base.authorizePlatformCapability(
		ctx, claims, target.capability, iamapplication.DeploymentModeSaaS,
	)
	if err != nil {
		return UsageAccessDecision{}, mapUsageBillingError(err)
	}
	decision, err := NewUsageAccessDecision(
		claims.Subject(), claims.Tenant(), uuid.Nil, runtime.environmentID,
		claims.TokenVersion(), claims.AuthorizationVersion(), authority.PlatformClock,
		request.Scope, authority.DeploymentMode, "saas_control", true,
		request.Permission, request.Operation,
	)
	if err != nil {
		return UsageAccessDecision{}, ErrUsageAccessUnavailable
	}
	return decision, nil
}

func mapUsageBillingError(err error) error {
	if errors.Is(err, ErrBillingAccessForbidden) {
		return ErrUsageAccessForbidden
	}
	return ErrUsageAccessUnavailable
}

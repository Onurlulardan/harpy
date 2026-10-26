package iam

import (
	"context"
	"errors"
	"time"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func (runtime *SupportAccessRuntime) Verify(
	ctx context.Context, request SupportAccessRequest,
) (SupportAccessDecision, error) {
	if runtime == nil || runtime.base == nil || ctx == nil || ctx.Err() != nil {
		return SupportAccessDecision{}, ErrSupportAccessUnavailable
	}
	target, targetValid := resolveSupportAccess(
		request.Scope, request.Permission, request.Operation,
	)
	wire, bearerValid := billingBearer(request.AuthorizationValues)
	now := runtime.base.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return SupportAccessDecision{}, ErrSupportAccessAuthentication
	}
	claims, err := runtime.base.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!tenantAccessHeadersMatch(ctx, runtime.base.pool,
			request.TenantIDValues, request.TenantSlugValues, claims.Tenant()) {
		return SupportAccessDecision{}, ErrSupportAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return SupportAccessDecision{}, ErrSupportAccessForbidden
	}
	if target.platform {
		return runtime.verifyPlatform(ctx, claims, request, target)
	}
	return runtime.verifyTenant(ctx, claims, request, target)
}

func (runtime *SupportAccessRuntime) verifyTenant(
	ctx context.Context, claims accesstoken.Claims,
	request SupportAccessRequest, target supportAccessTarget,
) (SupportAccessDecision, error) {
	authority, err := runtime.base.authorizeBilling(ctx, claims, billingPermissionTarget{
		resource: target.resource, action: target.action,
	})
	if err != nil {
		return SupportAccessDecision{}, mapSupportBillingError(err)
	}
	session, err := runtime.base.readConsumerSession(ctx, claims)
	if err != nil {
		if errors.Is(err, errConsumerSessionRejected) {
			return SupportAccessDecision{}, ErrSupportAccessForbidden
		}
		return SupportAccessDecision{}, ErrSupportAccessUnavailable
	}
	decision, err := NewSupportAccessDecision(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), authority.AuthorizationClock(),
		request.Scope, request.Permission, request.Operation,
		session.DisplayName, session.Email, session.TenantSlug, session.TenantSlug,
	)
	if err != nil {
		return SupportAccessDecision{}, ErrSupportAccessUnavailable
	}
	return decision, nil
}

func (runtime *SupportAccessRuntime) verifyPlatform(
	ctx context.Context, claims accesstoken.Claims,
	request SupportAccessRequest, target supportAccessTarget,
) (SupportAccessDecision, error) {
	authority, err := runtime.base.authorizePlatformCapability(
		ctx, claims, target.capability, iamapplication.DeploymentModeSaaS,
	)
	if err != nil {
		return SupportAccessDecision{}, mapSupportBillingError(err)
	}
	decision, err := NewSupportAccessDecision(
		claims.Subject(), uuid.Nil, claims.TokenVersion(), claims.AuthorizationVersion(),
		authority.PlatformClock, request.Scope, request.Permission, request.Operation,
		"", "", "", "",
	)
	if err != nil {
		return SupportAccessDecision{}, ErrSupportAccessUnavailable
	}
	return decision, nil
}

func mapSupportBillingError(err error) error {
	if errors.Is(err, ErrBillingAccessForbidden) {
		return ErrSupportAccessForbidden
	}
	return ErrSupportAccessUnavailable
}

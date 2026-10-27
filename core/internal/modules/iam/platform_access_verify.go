package iam

import (
	"context"
	"time"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *PlatformAccessRuntime) Verify(
	ctx context.Context,
	request PlatformAccessRequest,
) (PlatformAccessDecision, error) {
	if runtime == nil || runtime.base == nil || ctx == nil || ctx.Err() != nil {
		return PlatformAccessDecision{}, ErrPlatformAccessUnavailable
	}
	mode := iamapplication.DeploymentMode(request.ExpectedMode)
	if !iamapplication.ModeAllowsPlatform(mode, request.Capability) {
		return PlatformAccessDecision{}, ErrPlatformAccessAuthentication
	}
	wire, bearerValid := billingBearer(request.AuthorizationValues)
	now := runtime.base.clock().UTC().Truncate(time.Second)
	if !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return PlatformAccessDecision{}, ErrPlatformAccessAuthentication
	}
	claims, err := runtime.base.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.headersMatch(ctx, request, claims.Tenant()) {
		return PlatformAccessDecision{}, ErrPlatformAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return PlatformAccessDecision{}, ErrPlatformAccessForbidden
	}
	decision, err := runtime.base.authorizePlatformCapability(
		ctx, claims, request.Capability, mode)
	if err != nil || !decision.Allowed {
		if err == ErrBillingAccessForbidden {
			return PlatformAccessDecision{}, ErrPlatformAccessForbidden
		}
		return PlatformAccessDecision{}, ErrPlatformAccessUnavailable
	}
	result := platformAccessDecision(claims, decision, request)
	if request.RequireStepUp {
		stepUpAt, valid := customerProvisioningStepUp(
			claims, request.StepUpPurpose, now)
		if !valid {
			return PlatformAccessDecision{}, ErrPlatformAccessForbidden
		}
		result.StepUpAt = stepUpAt
	}
	if !result.Valid() {
		return PlatformAccessDecision{}, ErrPlatformAccessUnavailable
	}
	return result, nil
}

func platformAccessDecision(
	claims accesstoken.Claims,
	decision iamapplication.PlatformDecision,
	request PlatformAccessRequest,
) PlatformAccessDecision {
	return PlatformAccessDecision{
		SubjectID: claims.Subject(), TenantID: claims.Tenant(), SessionID: claims.Session(),
		DeploymentMode: decision.DeploymentMode, Capability: request.Capability,
		PlatformClock:        decision.PlatformClock,
		TokenVersion:         claims.TokenVersion(),
		AuthorizationVersion: claims.AuthorizationVersion(),
	}
}

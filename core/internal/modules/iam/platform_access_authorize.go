package iam

import (
	"context"
	"errors"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *BillingAccessRuntime) authorizePlatformCapability(
	ctx context.Context,
	claims accesstoken.Claims,
	capability string,
	mode iamapplication.DeploymentMode,
) (iamapplication.PlatformDecision, error) {
	if runtime == nil || runtime.platformIAM == nil || runtime.epoch == nil ||
		!iamapplication.ModeAllowsPlatform(mode, capability) {
		return iamapplication.PlatformDecision{}, ErrBillingAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.pool, claims)
	if err != nil {
		return iamapplication.PlatformDecision{}, ErrBillingAccessUnavailable
	}
	var result iamapplication.PlatformDecision
	err = runtime.coordinator.InTenant(ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if runtime.epoch.Assert(scoped, scope, claims.SecurityEpoch()) != nil {
				return ErrBillingAccessForbidden
			}
			current, readErr := runtime.platformIAM.ReadSessionContext(
				scoped, scope, iamapplication.PlatformSessionContextQuery{
					ActorID:                      iamapplication.PlatformID(claims.Subject()),
					ExpectedTokenVersion:         claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
				})
			purpose := iamapplication.TenantPurposeOnPremisePrimary
			if mode == iamapplication.DeploymentModeSaaS {
				purpose = iamapplication.TenantPurposeSaaSControl
			}
			if readErr != nil || current.DeploymentMode != string(mode) ||
				current.TenantPurpose != string(purpose) ||
				uuid.UUID(current.SubjectID) == uuid.Nil ||
				current.PlatformClock != claims.PlatformAuthorizationClock() {
				return ErrBillingAccessForbidden
			}
			decision, authorizeErr := runtime.platformIAM.Authorize(
				scoped, scope, iamapplication.PlatformEvaluateQuery{
					ActorID: iamapplication.PlatformID(claims.Subject()), Capability: capability,
					ExpectedTokenVersion:         claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
					ExpectedPlatformClock:        claims.PlatformAuthorizationClock(),
				})
			if authorizeErr != nil {
				return ErrBillingAccessUnavailable
			}
			if !decision.Allowed || decision.DeploymentMode != string(mode) ||
				decision.SubjectID != current.SubjectID {
				return ErrBillingAccessForbidden
			}
			result = decision
			return nil
		})
	if err != nil {
		if errors.Is(err, ErrBillingAccessForbidden) {
			return iamapplication.PlatformDecision{}, ErrBillingAccessForbidden
		}
		return iamapplication.PlatformDecision{}, ErrBillingAccessUnavailable
	}
	return result, nil
}

package iam

import (
	"context"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *BillingAccessRuntime) globalBillingDecision(
	ctx context.Context,
	claims accesstoken.Claims,
	request BillingAccessRequest,
	target billingPermissionTarget,
	platformDecision iamapplication.PlatformDecision,
) (BillingAccessDecision, error) {
	authority, err := iamapplication.NewPlatformBillingAuthority(
		platformDecision, claims.Subject(), target.capability,
		claims.AuthorizationVersion(),
	)
	if err != nil {
		return BillingAccessDecision{}, ErrBillingAccessForbidden
	}
	groupGeneration, billingGeneration, err :=
		runtime.pool.ResolveGlobalBillingGenerations(ctx)
	if err != nil {
		return BillingAccessDecision{}, ErrBillingAccessUnavailable
	}
	if target.command {
		capability, issueErr := iamapplication.BuildGlobalBillingCommandAccess(
			authority, request.Operation, request.Permission,
			groupGeneration, billingGeneration,
		)
		if issueErr != nil {
			return BillingAccessDecision{}, ErrBillingAccessUnavailable
		}
		return NewGlobalBillingCommandAccessDecision(
			claims.Subject(), claims.Tenant(), claims.TokenVersion(),
			claims.AuthorizationVersion(), platformDecision.PlatformClock, capability,
		)
	}
	capability, err := iamapplication.BuildGlobalBillingSnapshot(
		authority, request.Operation, request.Permission,
		groupGeneration, billingGeneration,
	)
	if err != nil {
		return BillingAccessDecision{}, ErrBillingAccessUnavailable
	}
	return NewGlobalBillingSnapshotAccessDecision(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), platformDecision.PlatformClock, capability,
	)
}

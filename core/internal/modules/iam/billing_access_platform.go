package iam

import (
	"context"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *BillingAccessRuntime) authorizePlatformBilling(
	ctx context.Context, claims accesstoken.Claims, target billingPermissionTarget,
) (iamapplication.PlatformDecision, error) {
	if !target.global {
		return iamapplication.PlatformDecision{}, ErrBillingAccessForbidden
	}
	return runtime.authorizePlatformCapability(
		ctx, claims, target.capability, iamapplication.DeploymentModeSaaS,
	)
}

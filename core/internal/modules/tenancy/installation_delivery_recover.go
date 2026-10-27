package tenancy

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

func (runtime *InstallationDeliveryRuntime) Recover(
	ctx context.Context,
	request onboardingapp.RecoveryRequest,
	callback onboardingapp.RecoveryCallback,
) error {
	if runtime == nil {
		return tenantapi.ErrOnboardingUnavailable
	}
	return onboardingapp.RecoverCredentialDelivery(ctx, runtime.recovery, request, callback)
}

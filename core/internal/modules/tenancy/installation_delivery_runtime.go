package tenancy

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type InstallationDeliveryRuntime struct {
	store    installationdelivery.Store
	recovery onboardingapp.RecoveryUnitOfWork
}

func NewInstallationDeliveryRuntime(
	store installationdelivery.Store,
	recovery onboardingapp.RecoveryUnitOfWork,
) (*InstallationDeliveryRuntime, error) {
	if store == nil || recovery == nil {
		return nil, api.ErrOnboardingUnavailable
	}
	return &InstallationDeliveryRuntime{store: store, recovery: recovery}, nil
}

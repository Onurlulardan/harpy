package tenancy

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (runtime *InstallationDeliveryRuntime) Claim(
	ctx context.Context,
	command installationdelivery.ClaimCommand,
) (installationdelivery.Delivery, error) {
	if runtime == nil {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return installationdelivery.Claim(ctx, runtime.store, command)
}

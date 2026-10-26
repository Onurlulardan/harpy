package tenancy

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (runtime *InstallationDeliveryRuntime) Stage(
	ctx context.Context,
	command installationdelivery.StageCommand,
) (installationdelivery.Delivery, error) {
	if runtime == nil {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return installationdelivery.Stage(ctx, runtime.store, command)
}

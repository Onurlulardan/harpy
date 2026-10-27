package tenancy

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (runtime *InstallationDeliveryRuntime) Acknowledge(
	ctx context.Context,
	command installationdelivery.AcknowledgeCommand,
) (installationdelivery.Delivery, error) {
	if runtime == nil {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return installationdelivery.Acknowledge(ctx, runtime.store, command)
}

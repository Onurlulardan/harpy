package tenancy

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (runtime *InstallationDeliveryRuntime) CompleteDelivery(
	ctx context.Context,
	command installationdelivery.StateCommand,
) (installationdelivery.Transition, error) {
	if runtime == nil {
		return installationdelivery.Transition{}, installationdelivery.ErrUnavailable
	}
	return installationdelivery.Complete(ctx, runtime.store, command)
}

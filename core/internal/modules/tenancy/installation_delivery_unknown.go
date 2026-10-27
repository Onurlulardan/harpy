package tenancy

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (runtime *InstallationDeliveryRuntime) MarkUnknown(
	ctx context.Context,
	command installationdelivery.UnknownCommand,
) (installationdelivery.Transition, error) {
	if runtime == nil {
		return installationdelivery.Transition{}, installationdelivery.ErrUnavailable
	}
	return installationdelivery.RecordUnknownOutcome(ctx, runtime.store, command)
}

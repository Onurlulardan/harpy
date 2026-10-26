package installationdelivery

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (storage *commandStore) Acknowledge(
	ctx context.Context, command installationdelivery.AcknowledgeCommand,
) (installationdelivery.Delivery, error) {
	runtime, bound, err := storage.open(ctx)
	if err != nil {
		return installationdelivery.Delivery{}, err
	}
	defer func() { _ = runtime.Rollback() }()
	result, err := bound.Acknowledge(ctx, command)
	if err != nil {
		return installationdelivery.Delivery{}, err
	}
	if err = commitCommand(ctx, runtime); err != nil {
		return installationdelivery.Delivery{}, err
	}
	return result, nil
}

package installationdelivery

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (storage *commandStore) Claim(
	ctx context.Context, command installationdelivery.ClaimCommand,
) (installationdelivery.Delivery, error) {
	runtime, bound, err := storage.open(ctx)
	if err != nil {
		return installationdelivery.Delivery{}, err
	}
	defer func() { _ = runtime.Rollback() }()
	result, err := bound.claim(ctx, command)
	if err != nil {
		return installationdelivery.Delivery{}, err
	}
	if err = commitCommand(ctx, runtime); err != nil {
		return installationdelivery.Delivery{}, err
	}
	return result, nil
}

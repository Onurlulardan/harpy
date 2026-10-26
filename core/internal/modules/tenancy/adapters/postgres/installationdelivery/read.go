package installationdelivery

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (storage *commandStore) Read(
	ctx context.Context, query installationdelivery.ReadQuery,
) (installationdelivery.Delivery, error) {
	runtime, bound, err := storage.open(ctx)
	if err != nil {
		return installationdelivery.Delivery{}, err
	}
	defer func() { _ = runtime.Rollback() }()
	result, err := bound.read(ctx, query)
	if err != nil {
		return installationdelivery.Delivery{}, err
	}
	if err = commitCommand(ctx, runtime); err != nil {
		return installationdelivery.Delivery{}, err
	}
	return result, nil
}

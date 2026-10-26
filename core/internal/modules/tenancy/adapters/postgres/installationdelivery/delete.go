package installationdelivery

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
)

func (storage *commandStore) DeleteSecret(
	ctx context.Context, command installationdelivery.StateCommand,
) (installationdelivery.Transition, error) {
	runtime, bound, err := storage.open(ctx)
	if err != nil {
		return installationdelivery.Transition{}, err
	}
	defer func() { _ = runtime.Rollback() }()
	result, err := bound.deleteSecret(ctx, command)
	if err != nil {
		return installationdelivery.Transition{}, err
	}
	if err = commitCommand(ctx, runtime); err != nil {
		return installationdelivery.Transition{}, err
	}
	return result, nil
}

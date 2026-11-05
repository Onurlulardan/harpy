package operator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func (runtime *deliveryRuntime) prepare(ctx context.Context) error {
	delivery, err := runtime.delivery.Read(ctx, runtime.identity.OperationID)
	if err = prepareDeliveryCredential(runtime.passwordPath(), delivery, err); err != nil {
		return err
	}
	switch delivery.State {
	case tenancy.InstallationDeliveryStateSecretDeleted, tenancy.InstallationDeliveryStateCompleted:
		_ = localinstall.RemoveRegular(runtime.statePath(claimTokenName))
		return localinstall.RemoveRegular(runtime.statePath(recoveryTokenName))
	default:
		return nil
	}
}

func (runtime *deliveryRuntime) prepareInitialCredential(ctx context.Context) error {
	delivery, err := runtime.delivery.Read(ctx, runtime.identity.OperationID)
	return prepareInitialCredential(runtime.passwordPath(), delivery, err)
}

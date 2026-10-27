package operator

import (
	"context"
	"errors"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func (runtime *deliveryRuntime) prepare(ctx context.Context) error {
	delivery, err := runtime.delivery.Read(ctx, runtime.identity.OperationID)
	if errors.Is(err, tenancy.ErrInstallationDeliveryUnavailable) {
		return ensurePassword(runtime.passwordPath())
	}
	if err != nil {
		return err
	}
	switch delivery.State {
	case tenancy.InstallationDeliveryStateReady:
		return ensurePassword(runtime.passwordPath())
	case tenancy.InstallationDeliveryStateClaimed, tenancy.InstallationDeliveryStateOutcomeUnknown:
		return nil
	case tenancy.InstallationDeliveryStateAcknowledged:
		return nil
	case tenancy.InstallationDeliveryStateSecretDeleted, tenancy.InstallationDeliveryStateCompleted:
		if _, statErr := os.Lstat(runtime.passwordPath()); statErr == nil {
			return installationsecret.ErrFenced
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
		_ = localinstall.RemoveRegular(runtime.statePath(claimTokenName))
		return localinstall.RemoveRegular(runtime.statePath(recoveryTokenName))
	default:
		return tenancy.ErrInstallationDeliveryUnavailable
	}
}

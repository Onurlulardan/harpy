package operator

import (
	"context"
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
)

func (runtime *deliveryRuntime) verify(ctx context.Context) error {
	delivery, err := runtime.delivery.Read(ctx, runtime.identity.OperationID)
	if err != nil {
		return err
	}
	if delivery.State != tenancy.InstallationDeliveryStateCompleted ||
		delivery.TenantID != runtime.identity.TenantID ||
		delivery.FirstOwnerUserID != runtime.identity.FirstUserID {
		return tenancy.ErrInstallationDeliveryUnavailable
	}
	if _, err = os.Lstat(runtime.passwordPath()); err == nil {
		return installationsecret.ErrFenced
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

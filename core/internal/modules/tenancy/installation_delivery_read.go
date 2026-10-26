package tenancy

import (
	"context"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
	"github.com/google/uuid"
)

func (runtime *InstallationDeliveryRuntime) Read(
	ctx context.Context, operationID uuid.UUID,
) (installationdelivery.Delivery, error) {
	if runtime == nil {
		return installationdelivery.Delivery{}, installationdelivery.ErrUnavailable
	}
	return installationdelivery.Read(ctx, runtime.store,
		installationdelivery.ReadQuery{OperationID: operationID})
}

package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *AirGapCoordinator) airGapResult(
	ctx context.Context,
	operation ports.OperationRecord,
) (api.AuthorityReceipt, error) {
	if err := service.base.applyCompletedRotation(ctx, operation); err != nil {
		return api.AuthorityReceipt{}, mapStoreError(err)
	}
	return receiptFrom(operation), nil
}

package billing

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) DecreaseJob(
	ctx context.Context,
	scope platformtx.JobTenantScope,
	ownerFactID uuid.UUID,
	bytes int64,
	receiptID uuid.UUID,
) error {
	if bytes <= 0 || !documentsUsageV7(ownerFactID) || !documentsUsageV7(receiptID) {
		return usageapp.ErrRequest
	}
	store, err := service.bindJobStore(scope)
	if err != nil {
		return err
	}
	commands, err := usageapp.NewCommands(store, service.nextID, service.clock)
	if err != nil {
		return err
	}
	key, err := usageapp.DocumentsStorageEntitlement()
	if err != nil {
		return err
	}
	periodID, err := commands.ResolvePeriod(ctx, key)
	if err != nil {
		return err
	}
	result, err := commands.AdjustGauge(ctx, usageapp.AdjustGaugeRequest{
		EntitlementKey: key, PeriodID: periodID,
		OwnerFactID: ownerFactID, Delta: -bytes,
	})
	if err != nil {
		return err
	}
	if !result.Allowed {
		return ErrDocumentsUsageDenied
	}
	return nil
}

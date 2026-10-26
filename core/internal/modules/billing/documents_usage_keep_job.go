package billing

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *DocumentsUsage) KeepJob(
	ctx context.Context,
	scope platformtx.JobTenantScope,
	ownerFactID, receiptID uuid.UUID,
) error {
	if !documentsUsageV7(ownerFactID) || !documentsUsageV7(receiptID) {
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
	key, keyErr := usageapp.DocumentsStorageEntitlement()
	if keyErr != nil {
		return keyErr
	}
	_, err = commands.ResolvePeriod(ctx, key)
	return err
}

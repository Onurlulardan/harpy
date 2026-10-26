package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) PutComponents(
	ctx context.Context,
	record ports.RecoveryComponentBatch,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validIdentity(record.RecoveryExecutionIdentity) {
		return ErrInvalid
	}
	if _, err := recoveryset.ParseComponentInventory(
		record.Components, true,
	); err != nil {
		return ErrInvalid
	}
	changed, err := service.store.PutRecoveryComponents(ctx, record)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}

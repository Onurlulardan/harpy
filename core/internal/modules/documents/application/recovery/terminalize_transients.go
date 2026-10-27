package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (service *Service) TerminalizeTransients(
	ctx context.Context,
	restoreOperationID uuid.UUID,
) (TransientDisposition, error) {
	if !service.usable(ctx) || restoreOperationID == uuid.Nil {
		return TransientDisposition{}, ErrInvalid
	}
	result, err := service.store.TerminalizeTransients(ctx, restoreOperationID)
	if err != nil || result.CancelledUploads < 0 ||
		result.ExpiredArchives < 0 || result.CancelledArchives < 0 {
		return TransientDisposition{}, ErrUnavailable
	}
	return result, nil
}

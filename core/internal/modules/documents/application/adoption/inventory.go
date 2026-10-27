package adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

func (application *Application) Inventory(ctx context.Context,
	request InventoryRequest,
) (Checkpoint, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validAuthority(request.Authority) ||
		!validCheckpoint(request.Checkpoint, request.Authority) {
		return Checkpoint{}, ErrInvalid
	}
	request.Checkpoint.State = CheckpointInventoried
	request.Checkpoint.UpdatedAt = application.clock.Now().UTC()
	if !foundation.ValidInstant(request.Checkpoint.UpdatedAt) {
		return Checkpoint{}, ErrUnavailable
	}
	result, _, err := application.store.Inventory(ctx, request)
	if err != nil {
		return Checkpoint{}, mapError(err)
	}
	if !validCheckpoint(result, request.Authority) ||
		result.State != CheckpointInventoried {
		return Checkpoint{}, ErrUnavailable
	}
	return result, nil
}

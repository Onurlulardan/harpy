package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (service *Service) BeginHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
) error {
	if !service.usable(ctx) || recoverySetID == uuid.Nil {
		return ErrInvalid
	}
	state, err := service.store.BeginHold(ctx, recoverySetID)
	if err != nil {
		return ErrUnavailable
	}
	if state != "HOLDING" && state != "COPYING" && state != "VERIFIED" {
		return ErrConflict
	}
	return nil
}

func (service *Service) usable(ctx context.Context) bool {
	return service != nil && service.store != nil &&
		ctx != nil && ctx.Err() == nil
}

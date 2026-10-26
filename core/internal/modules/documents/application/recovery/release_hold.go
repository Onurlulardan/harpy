package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (service *Service) ReleaseHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
	digest [32]byte,
) error {
	if !service.usable(ctx) || recoverySetID == uuid.Nil || digest == [32]byte{} {
		return ErrInvalid
	}
	changed, err := service.store.ReleaseHold(ctx, recoverySetID, digest)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}

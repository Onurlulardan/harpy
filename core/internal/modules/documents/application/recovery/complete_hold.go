package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (service *Service) CompleteHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
	digest [32]byte,
	count int64,
) error {
	if !service.usable(ctx) || recoverySetID == uuid.Nil ||
		digest == [32]byte{} || count < 0 {
		return ErrInvalid
	}
	changed, err := service.store.SealHold(ctx, recoverySetID, digest, count)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}

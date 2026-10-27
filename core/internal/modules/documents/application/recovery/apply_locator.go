package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (service *Service) ApplyLocator(
	ctx context.Context,
	remap LocatorRemap,
) error {
	if !service.usable(ctx) || remap.RestoreOperationID == uuid.Nil ||
		remap.RecoverySetID == uuid.Nil || !validInventoryItem(remap.Item) ||
		remap.TargetObjectKey != remap.Item.ObjectKey ||
		remap.TargetObjectVersion == "" || remap.TargetObjectGeneration < 1 {
		return ErrInvalid
	}
	changed, err := service.store.ApplyLocator(ctx, remap)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}

package recovery

import (
	"context"
	"regexp"

	"github.com/google/uuid"
)

var snapshotPattern = regexp.MustCompile(`^[0-9]+-[0-9A-F]+-[0-9]+$`)

func (service *Service) ReadInventoryAtSnapshot(
	ctx context.Context,
	recoverySetID uuid.UUID,
	snapshotID string,
) ([]InventoryItem, [32]byte, error) {
	if !service.usable(ctx) || recoverySetID == uuid.Nil ||
		!snapshotPattern.MatchString(snapshotID) {
		return nil, [32]byte{}, ErrInvalid
	}
	items, err := service.store.ReadInventoryAtSnapshot(
		ctx, recoverySetID, snapshotID,
	)
	if err != nil {
		return nil, [32]byte{}, ErrUnavailable
	}
	return normalizeInventory(items)
}

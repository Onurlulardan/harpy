package recovery

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	BeginHold(context.Context, uuid.UUID) (string, error)
	ListInventory(context.Context, uuid.UUID) ([]InventoryItem, error)
	ReadInventoryAtSnapshot(
		context.Context, uuid.UUID, string,
	) ([]InventoryItem, error)
	RecordCopy(context.Context, CopyReceipt) error
	SealHold(context.Context, uuid.UUID, [32]byte, int64) (bool, error)
	ReleaseHold(context.Context, uuid.UUID, [32]byte) (bool, error)
	FailHold(context.Context, uuid.UUID, string) (bool, error)
	ApplyLocator(context.Context, LocatorRemap) (bool, error)
	TerminalizeTransients(context.Context, uuid.UUID) (TransientDisposition, error)
}

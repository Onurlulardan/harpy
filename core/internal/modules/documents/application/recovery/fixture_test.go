package recovery

import (
	"context"

	"github.com/google/uuid"
)

type storeFixture struct {
	copyReceipt       CopyReceipt
	inventory         []InventoryItem
	snapshotInventory []InventoryItem
	sealed            bool
	released          bool
}

func (*storeFixture) BeginHold(context.Context, uuid.UUID) (string, error) {
	return "HOLDING", nil
}

func (fixture *storeFixture) ListInventory(
	context.Context, uuid.UUID,
) ([]InventoryItem, error) {
	return fixture.inventory, nil
}
func (fixture *storeFixture) ReadInventoryAtSnapshot(
	context.Context, uuid.UUID, string,
) ([]InventoryItem, error) {
	return fixture.snapshotInventory, nil
}

func (fixture *storeFixture) RecordCopy(_ context.Context, receipt CopyReceipt) error {
	fixture.copyReceipt = receipt
	return nil
}

func (fixture *storeFixture) SealHold(
	context.Context, uuid.UUID, [32]byte, int64,
) (bool, error) {
	return fixture.sealed, nil
}

func (fixture *storeFixture) ReleaseHold(
	context.Context, uuid.UUID, [32]byte,
) (bool, error) {
	return fixture.released, nil
}

func (*storeFixture) FailHold(context.Context, uuid.UUID, string) (bool, error) {
	return true, nil
}

func (*storeFixture) ApplyLocator(context.Context, LocatorRemap) (bool, error) {
	return true, nil
}

func (*storeFixture) TerminalizeTransients(
	context.Context, uuid.UUID,
) (TransientDisposition, error) {
	return TransientDisposition{}, nil
}

func newFixtureService(fixture *storeFixture) *Service {
	service, _ := New(fixture, Identity{
		EnvironmentID:   uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		TenantID:        uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		TenantSetSHA256: [32]byte{1},
	})
	return service
}

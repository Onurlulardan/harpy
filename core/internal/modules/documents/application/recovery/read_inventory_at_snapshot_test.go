package recovery

import (
	"testing"

	"github.com/google/uuid"
)

func TestReadInventoryAtSnapshotNormalizesAndValidates(t *testing.T) {
	first := validInventoryFixture(
		"10000000-0000-4000-8000-000000000001",
	)
	second := validInventoryFixture(
		"20000000-0000-4000-8000-000000000002",
	)
	fixture := &storeFixture{snapshotInventory: []InventoryItem{second, first}}
	service := newFixtureService(fixture)
	items, digest, err := service.ReadInventoryAtSnapshot(
		t.Context(), uuid.New(), "00000001-0000000A-1",
	)
	if err != nil || digest == [32]byte{} || items[0].VersionID != first.VersionID {
		t.Fatalf("snapshot inventory was not normalized: %#v %x %v", items, digest, err)
	}
	fixture.snapshotInventory = []InventoryItem{{VersionID: uuid.New()}}
	if _, _, err := service.ReadInventoryAtSnapshot(
		t.Context(), uuid.New(), "00000001-0000000A-1",
	); err != ErrUnavailable {
		t.Fatalf("malformed inventory error = %v", err)
	}
}

func TestReadInventoryAtSnapshotRejectsUnexportedIdentifier(t *testing.T) {
	service := newFixtureService(&storeFixture{})
	if _, _, err := service.ReadInventoryAtSnapshot(
		t.Context(), uuid.New(), "not-a-snapshot",
	); err == nil {
		t.Fatal("invalid exported snapshot identifier accepted")
	}
}

func validInventoryFixture(versionID string) InventoryItem {
	return InventoryItem{
		VersionID:        uuid.MustParse(versionID),
		ObjectKey:        "tenant/document",
		ObjectVersion:    "version-1",
		ObjectGeneration: 1,
		CiphertextSHA256: [32]byte{1},
		CiphertextBytes:  10,
		KeyID:            "key-1",
	}
}

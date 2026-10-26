package recovery

import (
	"strings"
	"testing"
)

func TestReadInventoryAtSnapshotUsesOwnerInventoryRoutine(t *testing.T) {
	if !strings.Contains(inventorySQL, "documents_recovery_inventory_v1") {
		t.Fatal("snapshot inventory bypasses the owner routine")
	}
}

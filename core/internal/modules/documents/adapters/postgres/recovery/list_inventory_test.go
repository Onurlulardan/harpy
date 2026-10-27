package recovery

import (
	"strings"
	"testing"
)

func TestListInventoryStatementUsesRecoveryInventoryRoutine(t *testing.T) {
	if !strings.Contains(inventorySQL, "documents_recovery_inventory_v1") {
		t.Fatal("inventory routine is not bound")
	}
}

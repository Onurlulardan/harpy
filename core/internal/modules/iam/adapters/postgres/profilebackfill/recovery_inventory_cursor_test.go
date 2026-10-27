package profilebackfill

import (
	"strings"
	"testing"
)

func TestRecoveryInventoryCursorStatementsAreFixedAndInjectionSafe(t *testing.T) {
	if !validRecoveryInventoryCursorName(recoveryInventoryCursorName) {
		t.Fatal("fixed recovery cursor name was rejected")
	}
	for _, invalid := range []string{"", recoveryInventoryCursorName + ";", "other"} {
		if validRecoveryInventoryCursorName(invalid) {
			t.Fatalf("invalid cursor %q was accepted", invalid)
		}
	}
	if strings.TrimSpace(fetchRecoveryInventorySQL) !=
		"FETCH FORWARD 64 FROM "+recoveryInventoryCursorName+";" {
		t.Fatalf("fetch SQL = %q", fetchRecoveryInventorySQL)
	}
	if strings.TrimSpace(closeRecoveryInventorySQL) !=
		"CLOSE "+recoveryInventoryCursorName+";" {
		t.Fatalf("close SQL = %q", closeRecoveryInventorySQL)
	}
}

func TestRecoveryInventoryOpenerHasNoClientCursorOrLimitInput(t *testing.T) {
	for _, marker := range []string{
		"iam_profile_recovery_inventory_open_v1(",
		"$1::uuid, $2::uuid, $3::uuid, $4::bigint, $5::bigint",
	} {
		if !strings.Contains(readRecoveryInventorySQL, marker) {
			t.Fatalf("inventory opener SQL missing %q", marker)
		}
	}
	for _, forbidden := range []string{"after_resource_id", "$6", "$7", "LIMIT"} {
		if strings.Contains(readRecoveryInventorySQL, forbidden) {
			t.Fatalf("inventory opener accepts forbidden %q", forbidden)
		}
	}
}

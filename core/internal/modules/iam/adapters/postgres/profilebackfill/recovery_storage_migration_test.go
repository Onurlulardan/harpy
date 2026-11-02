package profilebackfill

import (
	"strings"
	"testing"
)

var recoveryStorageMigrationNames = []string{
	"20260716020035_expand_iam_profile_recovery_lineage_v1",
	"20260716020036_guard_iam_profile_recovery_lineage_v1",
	"20260716020037_create_iam_profile_recovery_inventory_read_v1",
	"20260716020038_create_iam_profile_read_contract_v2",
	"20260716020039_create_iam_profile_write_contract_read_v2",
	"20260716020040_guard_iam_profile_recovery_writes_v1",
	"20260716020041_grant_iam_profile_recovery_runtime_v1",
}

func assertRecoveryMigrationMarker(t *testing.T, document, marker string) {
	t.Helper()
	if !strings.Contains(document, marker) {
		t.Fatalf("recovery migration missing %q", marker)
	}
}

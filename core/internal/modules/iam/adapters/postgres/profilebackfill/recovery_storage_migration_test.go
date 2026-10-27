package profilebackfill

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
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

func TestRecoveryStorageMigrationsStayPairedAndClosed(t *testing.T) {
	for _, name := range recoveryStorageMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readProfileBackfillMigration(t, "fresh", name)
			if fresh == "" {
				t.Fatal("fresh recovery migration is empty")
			}
			if strings.Count(fresh, "\n") > 200 {
				t.Fatal("recovery migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource([]byte(fresh)); err != nil {
				t.Fatalf("recovery migration source is invalid: %v", err)
			}
			upper := strings.ToUpper(fresh)
			for _, forbidden := range []string{
				"IF NOT EXISTS", "-- +GOOSE DOWN", "CREATE OR REPLACE FUNCTION",
				"CREATE TEMP", "CREATE TEMPORARY", "UNLOGGED", "PG_TEMP",
				"PUBLIC.", "ON DELETE CASCADE", "ON DELETE SET NULL",
			} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("recovery migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestRecoveryStorageBindsLinearImmutableLineage(t *testing.T) {
	expand := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[0])
	guard := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[1])
	for _, marker := range []string{
		"recovery_of_plan_id uuid", "recovery_root_plan_id uuid",
		"recovery_sequence bigint", "parent_plan_sha256 bytea",
		"inherited_target_count bigint", "remaining_source_count bigint",
		"remaining_source_count = 0 AND page_count = 0",
		"iam_profile_backfill_plans_recovery_parent_uidx",
		"iam_profile_backfill_plans_recovery_sequence_uidx",
	} {
		assertRecoveryMigrationMarker(t, expand, marker)
	}
	for _, marker := range []string{
		"parent.state", "v_parent_state IS DISTINCT FROM 'fenced'",
		"NEW.target_schema_id IS DISTINCT FROM v_parent_target_schema_id",
		"COALESCE(v_parent_sequence, 0) + 1",
		"Terminal IAM profile backfill evidence is immutable",
		"DEFERRABLE INITIALLY DEFERRED",
		"IAM profile recovery materialization is incomplete",
	} {
		assertRecoveryMigrationMarker(t, guard, marker)
	}
}

func TestRecoveryInventoryAndFencedContractsAreClosed(t *testing.T) {
	inventory := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[2])
	readContract := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[3])
	writeContract := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[4])
	writeGuard := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[5])
	acl := readProfileBackfillMigration(t, "fresh", recoveryStorageMigrationNames[6])
	for _, marker := range []string{
		"iam_profile_recovery_inventory_open_v1(",
		"v_cursor := 'iam_profile_recovery_inventory_v1'",
		"IF v_resource_kind = 'user' THEN",
		"FROM iam_users AS resource",
		"ELSIF v_resource_kind = 'organization' THEN",
		"FROM iam_organizations AS resource",
		"FROM iam_roles AS resource",
		"SELECT DISTINCT resource.schema_version_id",
		"OPEN v_cursor NO SCROLL FOR",
		"ORDER BY resource.id",
		"RETURN v_cursor;",
		"row_schema.state NOT IN ('active', 'retired')",
		"row_schema.version > v_source_version",
		"IAM profile recovery target document is invalid",
		"THEN 'target' ELSE 'source' END",
	} {
		assertRecoveryMigrationMarker(t, inventory, marker)
	}
	for _, forbidden := range []string{"LOCK TABLE", "FOR EXECUTE", "pg_catalog.format("} {
		if strings.Contains(inventory, forbidden) {
			t.Fatalf("recovery inventory cursor contains %q", forbidden)
		}
	}
	cursorContract := inventory[strings.Index(inventory, "OPEN v_cursor"):]
	if strings.Contains(cursorContract, "iam_profile_recovery_resources_v1") {
		t.Fatal("recovery cursor reads the union inventory view")
	}
	assertRecoveryMigrationMarker(t, readContract,
		"plan.state IN ('planned', 'running', 'verified', 'fenced')")
	for _, document := range []string{writeContract, writeGuard} {
		assertRecoveryMigrationMarker(t, document,
			"IAM profile domain is fenced pending recovery")
		assertRecoveryMigrationMarker(t, document, "ERRCODE = 'P5741'")
	}
	assertRecoveryMigrationMarker(t, writeContract,
		"plan.state IN ('planned', 'running')")
	assertRecoveryMigrationMarker(t, writeGuard,
		"plan.state IN ('planned','running')")
	for _, marker := range []string{
		"REVOKE EXECUTE ON FUNCTION iam_profile_read_contract_v1(name,text)",
		"REVOKE EXECUTE ON FUNCTION iam_profile_write_contract_read_v1(name,text)",
		"GRANT EXECUTE ON FUNCTION iam_profile_recovery_inventory_open_v1(",
		"GRANT EXECUTE ON FUNCTION iam_profile_read_contract_v2(name,text)",
		"GRANT EXECUTE ON FUNCTION iam_profile_write_contract_read_v2(name,text)",
		") TO harpy_runtime;",
	} {
		assertRecoveryMigrationMarker(t, acl, marker)
	}
}

func assertRecoveryMigrationMarker(t *testing.T, document, marker string) {
	t.Helper()
	if !strings.Contains(document, marker) {
		t.Fatalf("recovery migration missing %q", marker)
	}
}

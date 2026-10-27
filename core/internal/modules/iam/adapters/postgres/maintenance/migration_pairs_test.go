package maintenance

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

var maintenanceTenantMigrationNames = []string{
	"20260716014300_create_iam_maintenance_schedules",
	"20260716014301_create_iam_maintenance_checkpoints",
	"20260716014302_create_iam_maintenance_receipts",
	"20260716014303_create_iam_maintenance_observability",
	"20260716014304_create_iam_directory_key_evidence",
	"20260716014305_create_iam_profile_backfill_plans",
	"20260716014306_create_iam_profile_backfill_mappings",
	"20260716014307_create_iam_profile_backfill_pages",
	"20260716014308_create_iam_profile_backfill_errors",
	"20260716014309_allow_retired_iam_bcrypt_marker",
	"20260716014310_create_iam_maintenance_bootstrap_v1",
	"20260716014311_create_iam_maintenance_claim_assert_v1",
	"20260716014312_create_iam_maintenance_receipt_read_v1",
	"20260716014313_create_iam_maintenance_complete_recurring_v1",
	"20260716014320_create_iam_bcrypt_retire_job_v1",
	"20260716014321_create_iam_identifier_retire_job_v1",
	"20260716014322_create_iam_directory_tombstone_purge_job_v1",
	"20260716014323_create_iam_security_artifact_purge_job_v1",
	"20260716014330_create_iam_directory_repair_apply_v1",
	"20260716014331_create_iam_directory_repair_finish_v1",
	"20260716014332_create_iam_directory_checksum_finish_v1",
	"20260716014340_create_iam_profile_backfill_apply_v1",
	"20260716014341_create_iam_profile_backfill_error_v1",
	"20260716014342_create_iam_profile_backfill_receipt_v1",
	"20260716014343_create_iam_profile_backfill_finish_v1",
	"20260716014359_grant_iam_maintenance_runtime_v1",
	"20260716014658_guard_iam_directory_evidence_key_ids_v1",
	"20260716015920_create_iam_maintenance_directory_repair_page_v1",
	"20260716015921_create_iam_maintenance_directory_checksum_sources_v1",
	"20260716015922_create_iam_maintenance_profile_page_read_v1",
	"20260716015923_create_iam_maintenance_profile_mappings_read_v1",
	"20260716015924_create_iam_maintenance_profile_schema_read_v1",
	"20260716015925_create_iam_maintenance_profile_resources_read_v1",
	"20260716015926_grant_iam_maintenance_read_runtime_v1",
	"20260716015927_create_iam_maintenance_profile_resources_read_v2",
	"20260716015928_grant_iam_maintenance_profile_resources_read_v2",
	"20260716015939_create_iam_maintenance_profile_lineage_resources_read_v2",
	"20260716015940_create_iam_maintenance_profile_schema_read_v2",
	"20260716015941_create_iam_maintenance_profile_backfill_apply_v2",
	"20260716015942_create_iam_profile_backfill_activate_v2",
	"20260716015943_create_iam_maintenance_profile_backfill_finish_v2",
	"20260716015944_guard_iam_profile_backfill_source_writes_v2",
	"20260716015945_grant_iam_maintenance_profile_lineage_v2",
	"20260716015946_guard_iam_profile_snapshot_insert_v2",
	"20260716015950_create_iam_profile_backfill_draft_create_v2",
	"20260716015951_create_iam_profile_backfill_source_page_v2",
	"20260716015952_create_iam_profile_backfill_page_create_v2",
	"20260716015953_create_iam_profile_backfill_plan_create_v2",
	"20260716015954_grant_iam_profile_backfill_plan_create_v2",
	"20260716020012_create_iam_profile_worker_capabilities",
	"20260716020013_create_iam_profile_worker_capability_routines_v1",
	"20260716020014_create_iam_profile_worker_capability_consume_v1",
	"20260716020015_guard_iam_profile_live_write_v3",
	"20260716020016_create_iam_maintenance_profile_lineage_read_v3",
	"20260716020017_create_iam_maintenance_profile_backfill_apply_v3",
	"20260716020018_guard_iam_profile_worker_capability_commit_v1",
	"20260716020019_grant_iam_profile_live_backfill_v3",
}

var maintenancePlatformMigrationNames = []string{
	"20260716014607_guard_iam_directory_hmac_key_ids_v1",
	"20260716014608_create_iam_maintenance_directory_apply_v1",
	"20260716014609_create_iam_maintenance_directory_access_v1",
}

func TestMaintenanceMigrationOriginsAreExactPairs(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations")
	for scope, names := range map[string][]string{
		"platform": maintenancePlatformMigrationNames,
		"tenant":   maintenanceTenantMigrationNames,
	} {
		for _, name := range names {
			fresh, err := os.ReadFile(filepath.Join(
				root, scope, "fresh", name+".fresh.sql",
			))
			if err != nil {
				t.Fatalf("read fresh %s: %v", name, err)
			}
			adopted, err := os.ReadFile(filepath.Join(
				root, scope, "v1_adopted", name+".v1_adopted.sql",
			))
			if err != nil {
				t.Fatalf("read adopted %s: %v", name, err)
			}
			if !bytes.Equal(fresh, adopted) {
				t.Fatalf("fresh and V1-adopted maintenance migration differ: %s", name)
			}
		}
	}
}

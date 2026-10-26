package postgres

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

var iamSystemTransferMigrationNames = []string{
	"20260718030000_create_iam_system_record_exports_v1",
	"20260718030001_create_iam_system_import_sources_v1",
	"20260718030002_create_iam_system_record_imports_v1",
	"20260718030003_create_iam_system_transfer_progress_v1",
	"20260718030004_seed_iam_system_transfer_resources_v1",
	"20260718030005_seed_iam_system_transfer_permissions_v1",
	"20260718030006_create_iam_system_transfer_assert_v1",
	"20260718030007_create_iam_system_import_source_register_v1",
	"20260718030008_create_iam_system_record_export_submit_v1",
	"20260718030009_create_iam_system_record_import_submit_v1",
	"20260718030010_create_iam_system_record_export_read_v1",
	"20260718030011_create_iam_system_record_export_content_v1",
	"20260718030012_create_iam_system_record_export_cancel_v1",
	"20260718030013_create_iam_system_record_import_read_v1",
	"20260718030014_create_iam_system_record_import_cancel_v1",
	"20260718030015_create_iam_system_record_import_receipts_v1",
	"20260718030016_grant_iam_system_transfer_runtime_v1",
	"20260718030017_create_iam_system_import_worker_claim_v1",
	"20260718030018_create_iam_system_import_worker_receipts_v1",
	"20260718030019_create_iam_system_import_worker_commit_v1",
	"20260718030020_create_iam_system_import_worker_finish_v1",
	"20260718030021_grant_iam_system_import_worker_v1",
	"20260718030022_expand_iam_system_record_export_worker_v1",
	"20260718030023_create_iam_system_export_worker_load_v1",
	"20260718030024_create_iam_system_export_worker_claim_v1",
	"20260718030025_create_iam_system_export_worker_checkpoint_v1",
	"20260718030026_create_iam_system_export_worker_effect_v1",
	"20260718030027_create_iam_system_export_worker_complete_v1",
	"20260718030028_create_iam_system_export_worker_finish_v1",
	"20260718030029_create_iam_system_export_worker_unknown_v1",
	"20260718030030_grant_iam_system_export_worker_v1",
	"20260718030031_expand_iam_system_export_reconciliation_v1",
	"20260718030032_create_iam_system_export_reconcile_claim_v1",
	"20260718030033_create_iam_system_export_reconcile_sealed_v1",
	"20260718030034_create_iam_system_export_reconcile_missing_v1",
	"20260718030035_create_iam_system_export_reconcile_retry_v1",
	"20260718030036_create_iam_system_export_reconcile_park_v1",
	"20260718030037_grant_iam_system_export_reconcile_runtime_v1",
	"20260718030038_create_iam_system_export_reconcile_expired_v1",
	"20260718030039_grant_iam_system_export_reconcile_expired_v1",
	"20260718030050_create_iam_system_import_source_register_v2",
	"20260718030051_grant_iam_system_import_source_register_v2",
	"20260718030052_correct_iam_system_transfer_runtime_acl_v2",
	"20260718030100_add_iam_import_source_cleanup_v1",
	"20260718030101_create_iam_import_source_cleanup_claim_v1",
	"20260718030102_create_iam_import_source_cleanup_complete_v1",
	"20260718030103_create_iam_import_source_cleanup_retry_v1",
	"20260718030104_grant_iam_import_source_cleanup_v1",
	"20260721000000_create_iam_system_import_worker_claim_v2",
}

func TestIAMSystemTransferMigrationsStayPairedAndBounded(t *testing.T) {
	for _, name := range iamSystemTransferMigrationNames {
		t.Run(name, func(t *testing.T) {
			fresh := readIAMIdentityMigration(t, "fresh", name)
			adopted := readIAMIdentityMigration(t, "v1_adopted", name)
			if !bytes.Equal(fresh, adopted) {
				t.Fatal("fresh and v1_adopted transfer migration differs")
			}
			if bytes.Count(fresh, []byte("\n"))+1 > 200 {
				t.Fatal("IAM transfer migration exceeds 200 lines")
			}
			if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
				t.Fatalf("inspect IAM transfer migration: %v", err)
			}
			upper := strings.ToUpper(string(fresh))
			for _, forbidden := range []string{"CREATE TABLE IF NOT EXISTS",
				"CREATE INDEX IF NOT EXISTS", "CREATE SCHEMA IF NOT EXISTS",
				"ADD COLUMN IF NOT EXISTS", "CREATE OR REPLACE",
				"ON DELETE CASCADE", "FORMS_RECORD_", "-- +GOOSE DOWN"} {
				if strings.Contains(upper, forbidden) {
					t.Fatalf("migration contains forbidden token %q", forbidden)
				}
			}
		})
	}
}

func TestIAMSystemTransferStoragePinsClosedSecurityEvidence(t *testing.T) {
	joined := make([]byte, 0, 8192)
	for _, name := range iamSystemTransferMigrationNames {
		joined = append(joined, readIAMIdentityMigration(t, "fresh", name)...)
	}
	source := string(joined)
	for _, marker := range []string{"token_version", "authorization_version",
		"authorization_clock", "catalog_sha256", "identity_manifest_sha256",
		"idempotency_key_sha256", "source_sha256", "result_sha256",
		"principal_wait_slot BETWEEN 1 AND 5", "tenant_run_slot BETWEEN 1 AND 2",
		"'iam.users'", "'iam.roles'", "'iam.organizations'"} {
		if !strings.Contains(source, marker) {
			t.Fatalf("IAM transfer storage lacks %q", marker)
		}
	}
	for _, leaked := range []string{"email text", "name text", "mapped_values",
		"source_values", "password", "session_token"} {
		if strings.Contains(strings.ToLower(source), leaked) {
			t.Fatalf("IAM transfer storage leaks %q", leaked)
		}
	}
}

func TestIAMSystemTransferImportWorkerClaimV2ReturnsStoredPins(t *testing.T) {
	source := string(readIAMIdentityMigration(t, "fresh",
		"20260721000000_create_iam_system_import_worker_claim_v2"))
	for _, required := range []string{
		"iam_system_record_import_worker_claim_v1",
		"request.catalog_version_id", "request.catalog_sha256",
		"request.identity_manifest_sha256", "TO harpy_runtime",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("IAM import worker claim V2 lacks %q", required)
		}
	}
}

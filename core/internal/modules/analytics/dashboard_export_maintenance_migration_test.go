package analytics

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestDashboardExportMaintenanceMigrationsArePairedAndBounded(t *testing.T) {
	logicals := []string{
		"20260802010029_create_analytics_dashboard_export_maintenance_unknown_v1",
		"20260802010030_create_analytics_dashboard_export_maintenance_claim_v1",
		"20260802010031_create_analytics_dashboard_export_maintenance_lease_v1",
		"20260802010032_create_analytics_dashboard_export_maintenance_evidence_v1",
		"20260802010033_create_analytics_dashboard_export_maintenance_replay_v1",
		"20260802010034_create_analytics_dashboard_export_maintenance_resolve_v1",
		"20260802010035_fence_analytics_dashboard_export_delete_claim_v1",
		"20260802010036_fence_analytics_dashboard_export_delete_complete_v1",
	}
	for _, logical := range logicals {
		fresh := readDashboardExportPlatformMigration(t, "fresh", logical)
		adopted := readDashboardExportPlatformMigration(t, "v1_adopted", logical)
		if !bytes.Equal(fresh, adopted) {
			t.Fatalf("maintenance migration pair differs: %s", logical)
		}
		if bytes.Count(fresh, []byte{'\n'}) > 200 {
			t.Fatalf("maintenance migration exceeds 200 lines: %s", logical)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("maintenance migration %s is invalid: %v", logical, err)
		}
		if bytes.Contains(fresh, []byte("pg_catalog.extract(")) {
			t.Fatalf("maintenance migration schema-qualifies EXTRACT: %s", logical)
		}
		if bytes.Contains(fresh, []byte("<>")) {
			t.Fatalf("maintenance migration uses nullable raw inequality: %s", logical)
		}
	}
}

func TestDashboardExportMaintenanceWrappersLockPlatformFirst(t *testing.T) {
	tests := []struct{ logical, function, platform string }{
		{"20260802010030_create_analytics_dashboard_export_maintenance_claim_v1",
			"analytics_dashboard_export_maintenance_unknown_claim_v1",
			"platform_async_analytics_dashboard_export_maintenance_unknown_claim_v1"},
		{"20260802010031_create_analytics_dashboard_export_maintenance_lease_v1",
			"analytics_dashboard_export_maintenance_unknown_checkpoint_v1",
			"platform_async_analytics_dashboard_export_maintenance_unknown_checkpoint_v1"},
		{"20260802010031_create_analytics_dashboard_export_maintenance_lease_v1",
			"analytics_dashboard_export_maintenance_unknown_defer_v1",
			"platform_async_analytics_dashboard_export_maintenance_unknown_defer_v1"},
		{"20260802010034_create_analytics_dashboard_export_maintenance_resolve_v1",
			"analytics_dashboard_export_maintenance_unknown_resolve_v1",
			"platform_async_analytics_dashboard_export_maintenance_unknown_resolve_v1"},
		{"20260802010035_fence_analytics_dashboard_export_delete_claim_v1",
			"analytics_dashboard_export_delete_claim_v1",
			"platform_async_analytics_dashboard_export_maintenance_effect_owner_v1"},
		{"20260802010036_fence_analytics_dashboard_export_delete_complete_v1",
			"analytics_dashboard_export_delete_complete_v1",
			"platform_async_analytics_dashboard_export_maintenance_effect_owner_v1"},
	}
	for _, test := range tests {
		source := string(readDashboardExportPlatformMigration(t, "fresh", test.logical))
		body := dashboardExportPlatformFunctionBody(t, source, test.function)
		platformAt, tenantAt := strings.Index(body, test.platform), strings.Index(body, "FOR UPDATE")
		if platformAt < 0 || tenantAt < 0 || platformAt >= tenantAt {
			t.Fatalf("%s is not Platform-first: platform=%d tenant=%d",
				test.function, platformAt, tenantAt)
		}
	}
}

func TestDashboardExportMaintenanceResolutionFencesExactDelete(t *testing.T) {
	storage := strings.Join(strings.Fields(string(readDashboardExportPlatformMigration(t,
		"fresh", "20260802010029_create_analytics_dashboard_export_maintenance_unknown_v1"))), "")
	for _, required := range []string{
		"(reconcile_evidence_sha256ISNULLORpg_catalog.encode(reconcile_evidence_sha256,'hex')ISDISTINCTFROMpg_catalog.repeat('0',64))",
		"(delete_response_sha256ISNULLORpg_catalog.encode(delete_response_sha256,'hex')ISDISTINCTFROMpg_catalog.repeat('0',64))",
		"(absence_sha256ISNULLORpg_catalog.encode(absence_sha256,'hex')ISDISTINCTFROMpg_catalog.repeat('0',64))",
	} {
		if !strings.Contains(storage, required) {
			t.Fatalf("maintenance resolution storage accepts zero evidence digest: %s", required)
		}
	}
	evidence := string(readDashboardExportPlatformMigration(t, "fresh",
		"20260802010032_create_analytics_dashboard_export_maintenance_evidence_v1"))
	for _, required := range []string{
		"p_upload_completed_at IS NULL OR",
		"pg_catalog.isfinite(p_upload_completed_at)",
		"p_upload_completed_at<=p_reconcile_verified_at",
		"p_reconcile_sha256 IS DISTINCT FROM",
		"p_delete_response_sha256 IS DISTINCT FROM",
		"p_absence_sha256 IS DISTINCT FROM",
	} {
		if !strings.Contains(evidence, required) {
			t.Fatalf("maintenance evidence does not support exact missing/sealed shapes: %s",
				required)
		}
	}
	resolve := string(readDashboardExportPlatformMigration(t, "fresh",
		"20260802010034_create_analytics_dashboard_export_maintenance_resolve_v1"))
	for _, required := range []string{
		"delete_attempt_generation IS DISTINCT FROM p_delete_attempt_generation",
		"p_resolved_at<v_export.updated_at",
		"analytics_dashboard_export_maintenance_resolution_mode_v1",
	} {
		if !strings.Contains(resolve, required) {
			t.Fatalf("maintenance resolver lacks exact fence: %s", required)
		}
	}
	mode := string(readDashboardExportPlatformMigration(t, "fresh",
		"20260802010033_create_analytics_dashboard_export_maintenance_replay_v1"))
	for _, required := range []string{
		"p_export.delete_lease_owner=p_cleanup_lease_owner_id",
		"p_export.delete_lease_until=p_export.maintenance_cleanup_lease_until",
		"p_delete_verified_at>=p_local_state_at",
	} {
		if !strings.Contains(mode, required) {
			t.Fatalf("maintenance resolution mode lacks exact fence: %s", required)
		}
	}
}

func TestLegacyDeleteRuntimeCannotBypassPlatformOwnerFence(t *testing.T) {
	for _, logical := range []string{
		"20260802010035_fence_analytics_dashboard_export_delete_claim_v1",
		"20260802010036_fence_analytics_dashboard_export_delete_complete_v1",
	} {
		source := strings.Join(strings.Fields(string(
			readDashboardExportPlatformMigration(t, "fresh", logical))), "")
		for _, required := range []string{
			"platform_async_analytics_dashboard_export_maintenance_effect_owner_v1",
			"maintenance_unknown_job_idISNOTNULL",
			"FROMPUBLIC,harpy_runtime;",
			"TOharpy_runtime;",
		} {
			if !strings.Contains(source, required) {
				t.Fatalf("legacy delete fence %s lacks %s", logical, required)
			}
		}
	}
}

package async

import (
	"strings"
	"testing"
)

func TestAnalyticsDashboardExportRunningReceiptIdentityContract(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	identity := string(pairs["20260802020046_create_analytics_dashboard_export_running_receipt_identity_v1"]["fresh"])
	for _, token := range []string{
		"bfce9d42-fd8b-453a-a54e-d6f35a886868",
		"1bce05e3-fc4d-40a9-82e9-f125b12fa10a",
		"uuid_send(v_namespace)||pg_catalog.uuid_send(p_attempt_id)",
		"8e7a6d02-df6d-4327-b179-160296b50b8e",
		"p_effect_id::text", "decode('00','hex')",
		"ANALYTICS_EXPORT_DELETED", "ANALYTICS_EXPORT_EXPIRY_COMPLETE",
		"ANALYTICS_EXPORT_REJECTED", "ANALYTICS_EXPORT_CONFLICT",
		"ANALYTICS_EXPORT_UNAVAILABLE", "ANALYTICS_EXPORT_RETRY",
		"ANALYTICS_EXPORT_TERMINAL",
	} {
		if !strings.Contains(identity, token) {
			t.Fatalf("running receipt identity misses %q", token)
		}
	}
}

func TestAnalyticsDashboardExportRunningReceiptGuardContract(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	guard := string(pairs["20260802020047_create_analytics_dashboard_export_running_receipt_guard_v1"]["fresh"])
	for _, token := range []string{
		"maintenance_effect_owner_v1", "attempt.effect_request_sha256",
		"WHERE job.id=p_receipt.causation_id AND job.tenant_id=p_receipt.tenant_id\n    FOR UPDATE",
		"attempt.lease_generation=v_job.lease_generation FOR UPDATE",
		"maintenance_effect_id_v1", "maintenance_receipt_id_v2",
		"p_receipt.provider IS DISTINCT FROM 'analytics.dashboard-export'",
		"p_receipt.correlation_id IS DISTINCT FROM v_job.id::text",
		"p_receipt.causation_id IS DISTINCT FROM v_job.id",
		"p_receipt.provider_reference IS DISTINCT FROM v_job.payload_ref",
		"p_receipt.effect_result_ref IS DISTINCT FROM v_result_ref",
		"v_result_state NOT IN ('expired','cancelled','failed','deleted')",
		"WHEN 'ANALYTICS_EXPORT_CONFLICT' THEN 'conflict'",
		"WHEN 'ANALYTICS_EXPORT_RETRY' THEN 'transient'",
		"WHEN 'ANALYTICS_EXPORT_TERMINAL' THEN 'permanent'", "P6048",
	} {
		if !strings.Contains(guard, token) {
			t.Fatalf("running receipt guard misses %q", token)
		}
	}
	if strings.Contains(guard, "platform_async_assert_current_authority_v1") {
		t.Fatal("running maintenance receipt falls through to generic authority")
	}
}

func TestAnalyticsDashboardExportReceiptRouterPreservesOtherBranches(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	baseline := string(pairs["20260802020035_allow_analytics_dashboard_export_maintenance_receipt_v1"]["fresh"])
	router := string(pairs["20260802020048_route_analytics_dashboard_export_maintenance_receipt_v1"]["fresh"])
	maintenance := strings.Index(router, "platform_async_route_analytics_dashboard_export_maintenance_receipt_v1(\n            NEW")
	cutover := strings.Index(router, "platform_async_assert_cutover_effect_receipt_v1(NEW)")
	generic := strings.Index(router, "platform_async_assert_current_authority_v1(")
	if maintenance < 0 || cutover < 0 || generic < 0 ||
		maintenance > cutover || cutover > generic {
		t.Fatal("idempotency v10 branch order differs")
	}
	for _, token := range []string{
		"TG_OP IS DISTINCT FROM 'INSERT'", "TG_TABLE_SCHEMA IS DISTINCT FROM 'platform'",
		"TG_TABLE_NAME IS DISTINCT FROM 'idempotency_receipts'",
		"IF v_state='running'", "ELSIF v_state='dead'", "FOR UPDATE",
		"assert_analytics_dashboard_export_running_receipt_v1",
		"assert_analytics_dashboard_export_maintenance_receipt_v1",
		"EXISTS (SELECT 1 FROM platform.cutover_job_bindings",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CREATE OR REPLACE TRIGGER idempotency_receipts_authority_insert_guard",
	} {
		if !strings.Contains(router, token) {
			t.Fatalf("idempotency v10 router misses %q", token)
		}
	}
	if strings.Count(router, "CREATE OR REPLACE TRIGGER") != 1 ||
		strings.Contains(router, "jobs_authority_insert_guard") ||
		strings.Contains(router, "outbox_authority_insert_guard") ||
		strings.Contains(router, "dead_letters_authority_insert_guard") {
		t.Fatal("020048 rebound a non-idempotency trigger")
	}
	for trigger, table := range map[string]string{
		"jobs_authority_insert_guard":         "jobs",
		"outbox_authority_insert_guard":       "outbox",
		"dead_letters_authority_insert_guard": "dead_letters",
	} {
		binding := "CREATE OR REPLACE TRIGGER " + trigger +
			"\nBEFORE INSERT ON platform." + table + " FOR EACH ROW\n" +
			"EXECUTE FUNCTION platform.platform_async_validate_owner_insert_v9();"
		if !strings.Contains(baseline, binding) {
			t.Fatalf("%s is not bound to owner insert v9", trigger)
		}
	}
}

func TestAnalyticsDashboardExportReceiptReplayPrecedesStateRouting(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	replay := string(pairs["20260802020049_allow_analytics_dashboard_export_receipt_replay_v1"]["fresh"])
	existing := strings.Index(replay,
		"platform_async_assert_analytics_dashboard_export_receipt_replay_v1(\n        p_receipt")
	running := strings.Index(replay, "IF v_job.state='running'")
	if existing < 0 || running < 0 || existing > running {
		t.Fatal("exact receipt replay does not precede state routing")
	}
	for _, token := range []string{
		"SELECT job.* INTO v_job", "FOR UPDATE",
		"v_existing IS DISTINCT FROM p_receipt",
		"attempt.effect_receipt_id=p_receipt.receipt_id",
		"attempt.effect_id=p_receipt.effect_id",
		"attempt.effect_request_sha256=p_receipt.request_sha256",
		"maintenance_effect_id_v1", "maintenance_receipt_id_v2",
		"v_attempt.outcome NOT IN ('requeued','dead')",
		"v_attempt.outcome IS DISTINCT FROM 'succeeded'",
		"v_job.last_error_code='JOB_OUTCOME_UNKNOWN'", "P6048",
	} {
		if !strings.Contains(replay, token) {
			t.Fatalf("receipt replay guard misses %q", token)
		}
	}
}

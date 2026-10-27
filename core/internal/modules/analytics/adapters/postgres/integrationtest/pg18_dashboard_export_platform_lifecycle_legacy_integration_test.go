//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsDashboardExportLegacyRenderDenied(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	cases := []struct {
		name      string
		statement string
		arguments []any
	}{
		{"claim", `SELECT disposition FROM analytics_dashboard_export_claim_v1(
't_analytics_a',$1,$2,$3,$4)`, []any{analyticsExportID, analyticsExportJob,
			analyticsExportLeaseA, analyticsExportAt.Add(time.Second)}},
		{"effect_start", `SELECT analytics_dashboard_export_effect_start_v1(
't_analytics_a',$1,$2,1,'legacy/object','legacy-kek',$3,$4,$5,$6,$7,
1024,128,1,'text/csv','legacy.csv',$8)`, []any{analyticsExportID,
			analyticsExportLeaseA, bytes.Repeat([]byte{1}, 48),
			bytes.Repeat([]byte{2}, 12), bytes.Repeat([]byte{3}, 32),
			bytes.Repeat([]byte{4}, 32), bytes.Repeat([]byte{5}, 32),
			analyticsExportAt.Add(2 * time.Second)}},
		{"complete", `SELECT analytics_dashboard_export_complete_v1(
't_analytics_a',$1,$2,1,'legacy-version','legacy-provider',$3,$4)`,
			[]any{analyticsExportID, analyticsExportLeaseA,
				analyticsExportAt.Add(3 * time.Second),
				analyticsExportAt.Add(4 * time.Second)}},
		{"restart_missing", `SELECT analytics_dashboard_export_restart_missing_v1(
't_analytics_a',$1,$2,1,$3)`, []any{analyticsExportID, analyticsExportLeaseA,
			analyticsExportAt.Add(5 * time.Second)}},
		{"outcome_unknown", `SELECT analytics_dashboard_export_outcome_unknown_v1(
't_analytics_a',$1,$2,1,$3)`, []any{analyticsExportID, analyticsExportLeaseA,
			analyticsExportAt.Add(6 * time.Second)}},
		{"fail_render", `SELECT analytics_dashboard_export_fail_render_v1(
't_analytics_a',$1,$2,1,'LEGACY_FAILED',false,$3)`,
			[]any{analyticsExportID, analyticsExportLeaseA,
				analyticsExportAt.Add(7 * time.Second)}},
	}
	for _, testCase := range cases {
		before := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
			analyticsExportID, analyticsExportJob)
		transaction := analyticsExportTransaction(t, ctx, connection)
		_, err := transaction.ExecContext(ctx, testCase.statement, testCase.arguments...)
		_ = transaction.Rollback()
		requireAnalyticsPGCode(t, err, "42501", "legacy "+testCase.name)
		after := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
			analyticsExportID, analyticsExportJob)
		if before != after {
			t.Fatalf("legacy %s mutated tenant, Platform, quota, or capacity", testCase.name)
		}
	}
}

func analyticsDashboardExportLifecycleSnapshot(t *testing.T, ctx context.Context,
	database *sql.DB, exportID, jobID string,
) string {
	t.Helper()
	var snapshot string
	err := database.QueryRowContext(ctx, `SELECT jsonb_build_object(
'export',(SELECT to_jsonb(item) FROM t_analytics_a.analytics_dashboard_exports item
 WHERE item.export_id=$1),
'tenant_attempts',(SELECT COALESCE(jsonb_agg(to_jsonb(item) ORDER BY
 item.attempt_generation),'[]'::jsonb) FROM
 t_analytics_a.analytics_dashboard_export_platform_attempts item
 WHERE item.export_id=$1),
'tenant_receipts',(SELECT COALESCE(jsonb_agg(to_jsonb(item) ORDER BY
 item.attempt_generation),'[]'::jsonb) FROM
 t_analytics_a.analytics_dashboard_export_attempt_receipts item
 WHERE item.export_id=$1),
'job',(SELECT to_jsonb(item) FROM platform.jobs item WHERE item.id=$2),
'job_attempts',(SELECT COALESCE(jsonb_agg(to_jsonb(item) ORDER BY item.attempt),
 '[]'::jsonb) FROM platform.job_attempts item WHERE item.job_id=$2),
'platform_receipts',(SELECT COALESCE(jsonb_agg(to_jsonb(item) ORDER BY item.receipt_id),
 '[]'::jsonb) FROM platform.idempotency_receipts item WHERE item.causation_id=$2),
'quota',(SELECT to_jsonb(item) FROM platform.artifact_quota_counters item
 WHERE item.tenant_id=$3),
'tenant_capacity',(SELECT to_jsonb(item) FROM
 t_analytics_a.analytics_dashboard_export_capacity_counters item
 WHERE item.tenant_id=$3),
'principal_capacity',(SELECT to_jsonb(item) FROM
 t_analytics_a.analytics_dashboard_export_principal_counters item
 WHERE item.actor_id=$4),
'reservation',(SELECT to_jsonb(item) FROM
 t_analytics_a.analytics_dashboard_export_capacity_reservations item
 WHERE item.export_id=$1),
'capacity_receipts',(SELECT COALESCE(jsonb_agg(to_jsonb(item) ORDER BY
 item.transition_kind,item.attempt_generation),'[]'::jsonb) FROM
 t_analytics_a.analytics_dashboard_export_capacity_receipts item
 WHERE item.export_id=$1))::text`, exportID, jobID, analyticsTenantA,
		analyticsDashboardActor).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

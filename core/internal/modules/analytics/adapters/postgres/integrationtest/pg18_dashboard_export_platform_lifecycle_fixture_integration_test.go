//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const (
	analyticsFailureOperation    = "019b1000-0000-7000-8000-000000000090"
	analyticsFailureExport       = "019b1000-0000-7000-8000-000000000091"
	analyticsFailureJob          = "019b1000-0000-7000-8000-000000000092"
	analyticsFailureExpiryJob    = "019b1000-0000-7000-8000-000000000093"
	analyticsFailureLeaseA       = "019b1000-0000-7000-8000-000000000094"
	analyticsFailureEffect       = "019b1000-0000-7000-8000-000000000095"
	analyticsFailureLeaseB       = "019b1000-0000-7000-8000-000000000096"
	analyticsExpiryOperationV2   = "019b1000-0000-7000-8000-0000000000a0"
	analyticsExpiryExportV2      = "019b1000-0000-7000-8000-0000000000a1"
	analyticsExpiryRenderJobV2   = "019b1000-0000-7000-8000-0000000000a2"
	analyticsExpiryCleanupJobV2  = "019b1000-0000-7000-8000-0000000000a3"
	analyticsExpiryRenderLeaseV2 = "019b1000-0000-7000-8000-0000000000a4"
	analyticsExpiryRenderEffect  = "019b1000-0000-7000-8000-0000000000a5"
	analyticsExpiryReceipt       = "019b1000-0000-7000-8000-0000000000a6"
	analyticsDeleteLeaseV2       = "019b1000-0000-7000-8000-0000000000c1"
	analyticsDeleteAttemptV2     = "019b1000-0000-7000-8000-0000000000c2"
	analyticsExpiryLeaseV2       = "019b1000-0000-7000-8000-0000000000c3"
	analyticsExpiryAttemptV2     = "019b1000-0000-7000-8000-0000000000c4"
	analyticsUnknownOperationV2  = "019b1000-0000-7000-8000-0000000000d0"
	analyticsUnknownExportV2     = "019b1000-0000-7000-8000-0000000000d1"
	analyticsUnknownJobV2        = "019b1000-0000-7000-8000-0000000000d2"
	analyticsUnknownExpiryJobV2  = "019b1000-0000-7000-8000-0000000000d3"
	analyticsUnknownLeaseV2      = "019b1000-0000-7000-8000-0000000000d4"
	analyticsUnknownEffectV2     = "019b1000-0000-7000-8000-0000000000d5"
	analyticsUnknownCleanupV2    = "019b1000-0000-7000-8000-0000000000d6"
	analyticsUnknownCompetingV2  = "019b1000-0000-7000-8000-0000000000d7"
)

func prepareAnalyticsDashboardExportPlatformLifecycle(t *testing.T,
	ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	assertAnalyticsDashboardExportMaintenancePlannerSource(t)
	if _, err := database.ExecContext(ctx, `CREATE TABLE platform.cutover_job_bindings(
job_id uuid PRIMARY KEY,operation_id uuid,cutover_run_id uuid,
lease_manifest_sha256 text);
GRANT SELECT ON platform.cutover_job_bindings TO harpy_runtime`); err != nil {
		t.Fatalf("prepare Platform cutover binding fixture: %v", err)
	}
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsDashboardExportFoundationDocuments(t, origin))
	prepareAnalyticsDashboardExportAuthorityFixture(t, ctx, database)
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsDashboardExportProtocolDocuments(t, origin,
			"20260802020000", "20260802020024"))
	seededAt := time.Now().UTC().Add(-2 * time.Second).Truncate(time.Microsecond)
	for _, candidate := range []struct{ exportID, jobID string }{
		{analyticsExportID, analyticsExportJob},
		{analyticsFailureExport, analyticsFailureJob},
		{analyticsExpiryExportV2, analyticsExpiryRenderJobV2},
		{analyticsUnknownExportV2, analyticsUnknownJobV2},
	} {
		seedAnalyticsDashboardExportRenderJob(t, ctx, database,
			candidate.exportID, candidate.jobID, seededAt)
	}
	seedAnalyticsDashboardExportMaintenanceJob(t, ctx, database,
		analyticsExportID, analyticsExportDeleteJob, analyticsDeleteAttemptV2,
		analyticsDeleteLeaseV2, "analytics.dashboard.export.delete", 3, seededAt)
	seedAnalyticsDashboardExportMaintenanceJob(t, ctx, database,
		analyticsExpiryExportV2, analyticsExpiryCleanupJobV2,
		analyticsExpiryAttemptV2, analyticsExpiryLeaseV2,
		"analytics.dashboard.export.expire", 8, seededAt)
	applyAnalyticsDocuments(t, ctx, database, "",
		analyticsDashboardExportProtocolDocuments(t, origin,
			"20260802020025", "20260802020049"))
}

func prepareAnalyticsDashboardExportAuthorityFixture(t *testing.T,
	ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `CREATE FUNCTION
platform.platform_async_assert_worker_authority_v1(
p_scope_kind text,p_scope_id uuid,p_group text,p_group_generation bigint,
p_domain text,p_generation bigint,p_mode text) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog AS $f$ BEGIN
IF p_mode IS DISTINCT FROM 'v2' THEN
 RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='test mode is invalid';
END IF;
PERFORM platform.platform_async_assert_current_authority_v1(p_scope_kind,p_scope_id,
 p_group,p_group_generation,p_domain,p_generation); END $f$;
CREATE FUNCTION platform.platform_async_assert_job_row_authority_v1(
p_job platform.jobs) RETURNS void LANGUAGE plpgsql SECURITY DEFINER
SET search_path=pg_catalog AS $f$ BEGIN
PERFORM platform.platform_async_assert_current_authority_v1(p_job.scope_kind,
 p_job.scope_id,p_job.authority_group_id,p_job.authority_group_generation,
 p_job.authority_domain,p_job.cutover_generation); END $f$;
CREATE FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
uuid,uuid,text,uuid,text,text,bigint,text,timestamptz) RETURNS boolean
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $f$ SELECT false $f$;
GRANT EXECUTE ON FUNCTION platform.platform_async_assert_worker_authority_v1(
text,uuid,text,bigint,text,bigint,text) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
uuid,uuid,text,uuid,text,text,bigint,text,timestamptz) TO harpy_runtime`)
	if err != nil {
		t.Fatalf("prepare Platform authority fixture: %v", err)
	}
}

func seedAnalyticsDashboardExportRenderJob(t *testing.T, ctx context.Context,
	database *sql.DB, exportID, jobID string, createdAt time.Time,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO platform.jobs(
id,tenant_id,scope_kind,scope_id,authority_group_id,authority_group_generation,
authority_domain,cutover_generation,writer_origin,kind,payload_revision,payload_ref,
payload_sha256,dedupe_key,effect_class,effect_idempotency_key,priority,run_at,state,
attempts,max_attempts,lease_generation,created_at,updated_at)
VALUES($1,$2,'tenant',$2,'analytics',1,'analytics',1,'v2',
'analytics.dashboard.export.render',1,$3,encode(sha256(convert_to($3,'UTF8')),'hex'),
'analytics.dashboard.export.render:'||$3,'external',
'analytics.dashboard.export.render:'||$3,100,$4,'queued',0,3,0,$4,$4)`,
		jobID, analyticsTenantA, exportID, createdAt)
	if err != nil {
		t.Fatalf("seed Platform render job %s: %v", jobID, err)
	}
}

func seedAnalyticsDashboardExportMaintenanceJob(t *testing.T, ctx context.Context,
	database *sql.DB, exportID, jobID, attemptID, leaseOwner, kind string,
	maxAttempts int, createdAt time.Time,
) {
	t.Helper()
	effectAt := createdAt.Add(time.Second)
	request, requestSHA := analyticsDashboardExportMaintenanceRequest(t,
		exportID, jobID, attemptID, kind, effectAt)
	_, err := database.ExecContext(ctx, `WITH inserted_job AS (
INSERT INTO platform.jobs(id,tenant_id,scope_kind,scope_id,authority_group_id,
authority_group_generation,authority_domain,cutover_generation,writer_origin,kind,
payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
effect_idempotency_key,priority,run_at,state,attempts,max_attempts,lease_owner_id,
lease_generation,lease_heartbeat_at,lease_until,effect_started_at,created_at,updated_at)
VALUES($1,$2,'tenant',$2,'analytics',1,'analytics',1,'v2',$3,1,$4,
encode(sha256(convert_to($4,'UTF8')),'hex'),$3||':'||$4,'external',$3||':'||$4,
100,$5::timestamptz,'running',1,$6,$7,1,$8::timestamptz,
$8::timestamptz+interval '5 minutes',$8::timestamptz,$5::timestamptz,
$8::timestamptz) RETURNING id)
INSERT INTO platform.job_attempts(attempt_id,job_id,scope_kind,scope_id,attempt,
lease_generation,lease_owner_id,started_at,effect_started_at,effect_id,
effect_request_sha256)
SELECT $9,$1,'tenant',$2,1,1,$7,$5::timestamptz,$8::timestamptz,$10,$11
FROM inserted_job`,
		jobID, analyticsTenantA, kind, exportID, createdAt, maxAttempts,
		leaseOwner, effectAt, attemptID, request.EffectID, requestSHA)
	if err != nil {
		t.Fatalf("seed Platform maintenance job %s: %v", jobID, err)
	}
}

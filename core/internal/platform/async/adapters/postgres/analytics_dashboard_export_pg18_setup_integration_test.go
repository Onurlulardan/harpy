//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const analyticsDashboardExportPG18Fixture = `
ALTER TABLE platform.tenants
    ADD COLUMN IF NOT EXISTS schema_name name,
    ADD COLUMN IF NOT EXISTS status text,
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
UPDATE platform.tenants SET schema_name='tenant_a',status='ACTIVE';
CREATE SCHEMA tenant_a AUTHORIZATION harpy_owner;
INSERT INTO platform.cutover_groups(
    scope_kind,scope_id,group_id,state,writer,generation
) VALUES ('tenant','11111111-1111-4111-8111-111111111111',
    'analytics','v2','v2',1);
INSERT INTO platform.domain_cutovers(
    scope_kind,scope_id,domain,state,writer,generation,changed_at
) VALUES ('tenant','11111111-1111-4111-8111-111111111111',
    'analytics','v2','v2',1,clock_timestamp());
INSERT INTO platform.cutover_group_members(scope_kind,scope_id,group_id,domain)
VALUES ('tenant','11111111-1111-4111-8111-111111111111','analytics','analytics');
INSERT INTO platform.artifact_quota_counters(tenant_id)
VALUES ('11111111-1111-4111-8111-111111111111');
CREATE TABLE platform.cutover_job_bindings(
    job_id uuid PRIMARY KEY,operation_id uuid,cutover_run_id uuid,
    lease_manifest_sha256 text
);
GRANT SELECT ON platform.cutover_job_bindings TO harpy_runtime;
CREATE FUNCTION platform.platform_async_assert_worker_authority_v1(
    p_scope_kind text,p_scope_id uuid,p_group text,p_group_generation bigint,
    p_domain text,p_generation bigint,p_mode text
) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ BEGIN
    IF p_mode IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='test mode is invalid';
    END IF;
    PERFORM platform.platform_async_assert_current_authority_v1(
        p_scope_kind,p_scope_id,p_group,p_group_generation,p_domain,p_generation);
END $f$;
CREATE OR REPLACE FUNCTION platform.platform_async_assert_job_row_authority_v1(
    p_job platform.jobs
) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ BEGIN
    PERFORM platform.platform_async_assert_current_authority_v1(
        p_job.scope_kind,p_job.scope_id,p_job.authority_group_id,
        p_job.authority_group_generation,p_job.authority_domain,
        p_job.cutover_generation);
END $f$;
CREATE FUNCTION platform.platform_async_cutover_shadow_operation_active_v1(
    uuid,uuid,text,uuid,text,text,bigint,text,timestamptz
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT false $f$;
GRANT EXECUTE ON FUNCTION platform.platform_async_assert_worker_authority_v1(
    text,uuid,text,bigint,text,bigint,text) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_cutover_shadow_operation_active_v1(
        uuid,uuid,text,uuid,text,text,bigint,text,timestamptz)
TO harpy_runtime;`

func prepareAnalyticsDashboardExportPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
) *pgxpool.Conn {
	t.Helper()
	fixtures := []struct{ name, statement string }{
		{"base", analyticsDashboardExportPG18Fixture},
		{"worker wrappers", analyticsDashboardExportWorkerWrappersPG18},
		{"recovery wrappers", analyticsDashboardExportRecoveryWrappersPG18},
	}
	for _, fixture := range fixtures {
		if _, err := database.ExecContext(ctx, fixture.statement); err != nil {
			t.Fatalf("prepare Analytics dashboard export %s: %v", fixture.name, err)
		}
	}
	return openAsyncRuntimeRole(t, ctx, database)
}

func seedAnalyticsDashboardExportJob(
	t *testing.T, ctx context.Context, runtime *pgxpool.Conn,
	exportID, jobID uuid.UUID, createdAt time.Time,
) {
	t.Helper()
	_, err := runtime.Exec(ctx, `INSERT INTO platform.jobs(
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,
    writer_origin,kind,payload_revision,payload_ref,payload_sha256,dedupe_key,
    effect_class,effect_idempotency_key,priority,run_at,state,attempts,
    max_attempts,lease_generation
) VALUES ($1,$2,'tenant',$2,'analytics',1,'analytics',1,'v2',
    'analytics.dashboard.export.render',1,$3,
    encode(sha256(convert_to($3,'UTF8')),'hex'),
    'analytics.dashboard.export.render:'||$3,'external',
    'analytics.dashboard.export.render:'||$3,10,$4,'queued',0,3,0)`,
		jobID, asyncTenantA, exportID.String(), createdAt)
	if err != nil {
		t.Fatalf("seed Analytics dashboard export job: %v", err)
	}
}

type analyticsDashboardExportClaim struct {
	state      string
	attempt    pgtype.UUID
	claimed    pgtype.Timestamptz
	leaseUntil pgtype.Timestamptz
}

func claimAnalyticsDashboardExport(
	t *testing.T, ctx context.Context, runtime *pgxpool.Conn,
	exportID, jobID, leaseID uuid.UUID, generation int64,
) analyticsDashboardExportClaim {
	t.Helper()
	var claim analyticsDashboardExportClaim
	if err := runtime.QueryRow(ctx, `SELECT * FROM tenant_a.test_analytics_export_claim(
        $1,$2,$3,$4)`, exportID, jobID, generation, leaseID).Scan(
		&claim.state, &claim.attempt, &claim.claimed, &claim.leaseUntil,
	); err != nil {
		t.Fatalf("claim Analytics dashboard export: %v", err)
	}
	return claim
}

//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const analyticsDashboardExportMaintenanceWrappersPG18 = `
CREATE TABLE tenant_a.test_analytics_export_maintenance_mutations(
    export_id uuid PRIMARY KEY,kind text NOT NULL
);
GRANT SELECT,INSERT ON tenant_a.test_analytics_export_maintenance_mutations
TO harpy_runtime;
CREATE FUNCTION tenant_a.test_analytics_export_maintenance_effect_owner(
    p_export uuid,p_job uuid,p_effect uuid,p_observed timestamptz
) RETURNS text LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS $f$ SELECT platform.platform_async_analytics_dashboard_export_maintenance_effect_owner_v1(
    '11111111-1111-4111-8111-111111111111',p_export,p_job,p_effect,p_observed) $f$;
GRANT EXECUTE ON FUNCTION tenant_a.test_analytics_export_maintenance_effect_owner(
    uuid,uuid,uuid,timestamptz) TO harpy_runtime;`

type analyticsMaintenanceActive struct {
	kind                       string
	exportID, jobID, attemptID uuid.UUID
	leaseID, effectID          uuid.UUID
	claim                      JobClaim
	permit                     ExternalEffectPermit
	repository                 *workerRepository
	authority                  JobAuthority
}

func prepareAnalyticsMaintenancePG18(
	t *testing.T, fixture analyticsDashboardExportPG18,
) {
	t.Helper()
	if _, err := fixture.database.ExecContext(fixture.ctx,
		analyticsDashboardExportMaintenanceWrappersPG18); err != nil {
		t.Fatalf("prepare Analytics maintenance wrappers: %v", err)
	}
}

func createAnalyticsMaintenanceActive(
	t *testing.T, fixture analyticsDashboardExportPG18, kind string,
) analyticsMaintenanceActive {
	t.Helper()
	maximum, operation, namespace := 3, "delete",
		uuid.MustParse("bfce9d42-fd8b-453a-a54e-d6f35a886868")
	if kind == "analytics.dashboard.export.expire" {
		maximum, operation = 8, "expire"
		namespace = uuid.MustParse("1bce05e3-fc4d-40a9-82e9-f125b12fa10a")
	}
	evidence := analyticsMaintenanceActive{
		kind: kind, exportID: analyticsExportV7(t), jobID: analyticsExportV7(t),
		leaseID: analyticsExportV7(t),
	}
	_, err := fixture.runtime.Exec(fixture.ctx, `INSERT INTO platform.jobs(
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,
    writer_origin,kind,payload_revision,payload_ref,payload_sha256,dedupe_key,
    effect_class,effect_idempotency_key,priority,run_at,state,attempts,
    max_attempts,lease_generation
) VALUES ($1,$2,'tenant',$2,'analytics',1,'analytics',1,'v2',$3,1,$4,
    encode(sha256(convert_to($4,'UTF8')),'hex'),$3||':'||$4,
    'external',$3||':'||$4,10,$5,'queued',0,$6,0)`, evidence.jobID,
		asyncTenantA, kind, evidence.exportID.String(), time.Now().UTC().Add(-time.Second),
		maximum)
	if err != nil {
		t.Fatalf("seed Analytics maintenance job: %v", err)
	}
	identity := &workerIdentity{marker: 1}
	authority, err := newJobAuthority(identity, "tenant", uuid.MustParse(asyncTenantA),
		"analytics", 1, "analytics", 1)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := jobsv1.NewKindPolicy(kind, 1, "analytics", jobsv1.EffectExternal, maximum)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	workerRepo := &workerRepository{
		database: fixture.runtime, cleanup: 5 * time.Second, identity: identity,
	}
	claims, err := workerRepo.ClaimJobs(fixture.ctx, authority, evidence.leaseID, 1, catalog)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim Analytics maintenance job: claims=%d err=%v", len(claims), err)
	}
	evidence.claim, evidence.attemptID = claims[0], claims[0].AttemptID()
	evidence.effectID = uuid.NewSHA1(namespace, evidence.attemptID[:])
	tx, err := fixture.runtime.BeginTx(fixture.ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(fixture.ctx)
	bound := &repository{
		database:  lifecycleBoundPGTx{transaction: tx},
		authority: lifecycleAuthorityStamp(authority),
	}
	spec, err := NewExternalEffectRequestSpec(evidence.effectID,
		"analytics.dashboard-export", operation, evidence.jobID.String(), evidence.jobID)
	if err != nil {
		t.Fatal(err)
	}
	evidence.permit, err = bound.MarkExternalEffectStarted(fixture.ctx, evidence.claim, spec)
	if err != nil {
		t.Fatalf("start Analytics maintenance effect: %v", err)
	}
	if err = tx.Commit(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	evidence.repository, evidence.authority = workerRepo, authority
	return evidence
}

func analyticsMaintenanceJobSnapshot(
	t *testing.T, ctx context.Context, fixture analyticsDashboardExportPG18, jobID uuid.UUID,
) string {
	t.Helper()
	var snapshot string
	err := fixture.database.QueryRowContext(ctx, `SELECT jsonb_build_object(
    'job',to_jsonb(job),'attempt',to_jsonb(attempt))::text
FROM platform.jobs job JOIN platform.job_attempts attempt ON attempt.job_id=job.id
WHERE job.id=$1 AND attempt.attempt=job.attempts`, jobID).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func setAnalyticsMaintenanceTerminal(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsMaintenanceActive, state string,
) {
	t.Helper()
	completed := time.Now().UTC()
	outcome, code, retain := "succeeded", any(nil), completed.Add(90*24*time.Hour)
	if state == "dead" {
		outcome, code, retain = "dead", "JOB_OUTCOME_UNKNOWN", completed.Add(365*24*time.Hour)
	}
	tx, err := fixture.database.BeginTx(fixture.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	for _, table := range []string{"platform.job_attempts", "platform.jobs"} {
		if _, err = tx.ExecContext(fixture.ctx, "ALTER TABLE "+table+" DISABLE TRIGGER USER"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = tx.ExecContext(fixture.ctx, `UPDATE platform.job_attempts
SET finished_at=$2,outcome=$3,error_code=$4 WHERE attempt_id=$1`,
		evidence.attemptID, completed, outcome, code); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.ExecContext(fixture.ctx, `UPDATE platform.jobs SET state=$2,
lease_owner_id=NULL,lease_heartbeat_at=NULL,lease_until=NULL,last_error_code=$3,
completed_at=$4,updated_at=$4,retain_until=$5 WHERE id=$1`, evidence.jobID,
		state, code, completed, retain); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"platform.job_attempts", "platform.jobs"} {
		if _, err = tx.ExecContext(fixture.ctx, "ALTER TABLE "+table+" ENABLE TRIGGER USER"); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type analyticsDashboardExportPG18 struct {
	ctx      context.Context
	database *sql.DB
	runtime  *pgxpool.Conn
}

func startAnalyticsDashboardExportPG18(t *testing.T) analyticsDashboardExportPG18 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	t.Cleanup(cancel)
	database := startAdoptionPG18(t, ctx, "analytics_export_behavior")
	applyAnalyticsDashboardExportPG18Migrations(t, ctx, database, "fresh")
	runtime := prepareAnalyticsDashboardExportPG18(t, ctx, database)
	return analyticsDashboardExportPG18{ctx: ctx, database: database, runtime: runtime}
}

func analyticsExportV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func analyticsExportDigest(marker byte) []byte {
	value := make([]byte, 32)
	value[0] = marker
	return value
}

func analyticsExportAttemptID(t *testing.T, claim analyticsDashboardExportClaim) uuid.UUID {
	t.Helper()
	if !claim.attempt.Valid {
		t.Fatal("Analytics dashboard export claim has no attempt")
	}
	return uuid.UUID(claim.attempt.Bytes)
}

func startAnalyticsDashboardExportEffect(
	t *testing.T, fixture analyticsDashboardExportPG18,
	exportID, jobID, attemptID, leaseID, effectID uuid.UUID,
	request []byte,
) time.Time {
	t.Helper()
	startedAt := time.Now().UTC()
	var stored time.Time
	err := fixture.runtime.QueryRow(fixture.ctx,
		`SELECT tenant_a.test_analytics_export_effect_start(
            $1,$2,$3,1,$4,$5,$6,$7)`,
		exportID, jobID, attemptID, leaseID, effectID, request, startedAt,
	).Scan(&stored)
	if err != nil {
		t.Fatalf("start Analytics dashboard export effect: %v", err)
	}
	return stored
}

func setAnalyticsDashboardExportLease(
	t *testing.T, fixture analyticsDashboardExportPG18,
	jobID uuid.UUID, leaseUntil time.Time,
) {
	t.Helper()
	transaction, err := fixture.database.BeginTx(fixture.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(fixture.ctx,
		"ALTER TABLE platform.jobs DISABLE TRIGGER USER"); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(fixture.ctx,
		`UPDATE platform.jobs SET lease_heartbeat_at=$2::timestamptz-interval '60 seconds',
			lease_until=$2::timestamptz WHERE id=$1`, jobID, leaseUntil); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(fixture.ctx,
		"ALTER TABLE platform.jobs ENABLE TRIGGER USER"); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func analyticsDashboardExportQuota(
	t *testing.T, fixture analyticsDashboardExportPG18,
) (reserved int64, revision int64) {
	t.Helper()
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT
    reserved_ciphertext_bytes,revision FROM platform.artifact_quota_counters
    WHERE tenant_id=$1`, asyncTenantA).Scan(&reserved, &revision)
	if err != nil {
		t.Fatal(err)
	}
	return reserved, revision
}

func analyticsDashboardExportSnapshot(
	t *testing.T, fixture analyticsDashboardExportPG18, jobID uuid.UUID,
) string {
	t.Helper()
	var snapshot string
	err := fixture.database.QueryRowContext(fixture.ctx, `SELECT jsonb_build_object(
    'job',to_jsonb(job),'attempt',to_jsonb(attempt),'quota',to_jsonb(quota)
)::text FROM platform.jobs AS job
JOIN platform.job_attempts AS attempt ON attempt.job_id=job.id
JOIN platform.artifact_quota_counters AS quota ON quota.tenant_id=job.tenant_id
WHERE job.id=$1 AND attempt.attempt=job.attempts`, jobID).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func assertGenericRecoverySkipsAnalyticsExport(
	t *testing.T, fixture analyticsDashboardExportPG18, jobID uuid.UUID,
) {
	t.Helper()
	before := analyticsDashboardExportSnapshot(t, fixture, jobID)
	identity := &workerIdentity{marker: 1}
	repository := &workerRepository{
		database: fixture.runtime, cleanup: 5 * time.Second, identity: identity,
	}
	authority, err := newJobAuthority(
		identity, "tenant", uuid.MustParse(asyncTenantA), "analytics", 1, "analytics", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	result, err := repository.RecoverExpiredJobs(fixture.ctx, authority, 10)
	after := analyticsDashboardExportSnapshot(t, fixture, jobID)
	if err != nil || result.Total() != 0 || before != after {
		t.Fatalf("generic recovery result=%#v changed=%v err=%v",
			result, before != after, err)
	}
}

func analyticsExportSQLState(err error) string {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code
	}
	return ""
}

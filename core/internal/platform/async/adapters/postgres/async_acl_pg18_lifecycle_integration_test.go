//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	asyncACLJobID     = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"
	asyncACLAttemptID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2"
	asyncACLLeaseID   = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3"
)

func assertAsyncACLJobLifecycle(
	t *testing.T, ctx context.Context, runtime *pgxpool.Conn,
) {
	t.Helper()
	if _, err := runtime.Exec(ctx, `SELECT platform.platform_async_assert_current_authority_v1(
        'tenant',$1,'operations',1,'workflow',1
    )`, asyncTenantA); err != nil {
		t.Fatalf("runtime authority assertion: %v", err)
	}
	if _, err := runtime.Exec(ctx, `
INSERT INTO platform.jobs (
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,
    writer_origin,kind,payload_revision,payload_ref,payload_sha256,dedupe_key,
    effect_class,priority,run_at,state,attempts,max_attempts,lease_generation
) VALUES (
    $1,$2,'tenant',$2,'operations',1,'workflow',1,
    'v2','workflow.acl',1,'job://acl',repeat('a',64),'acl-job',
    'database_transactional',0,clock_timestamp(),'queued',0,1,0
)`, asyncACLJobID, asyncTenantA); err != nil {
		t.Fatalf("runtime job insert through owner guard: %v", err)
	}
	if _, err := runtime.Exec(ctx, `
WITH stamp AS (
    SELECT GREATEST(clock_timestamp(),updated_at+interval '1 microsecond') AS at
    FROM platform.jobs WHERE id=$1
)
UPDATE platform.jobs SET state='running',attempts=1,lease_generation=1,
    lease_owner_id=$2,lease_heartbeat_at=stamp.at,
    lease_until=stamp.at+interval '60 seconds',updated_at=stamp.at
FROM stamp WHERE id=$1`, asyncACLJobID, asyncACLLeaseID); err != nil {
		t.Fatalf("runtime job claim through nested validator: %v", err)
	}
	if _, err := runtime.Exec(ctx, `
INSERT INTO platform.job_attempts (
    attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
    lease_owner_id,started_at
)
SELECT $1,id,scope_kind,scope_id,attempts,lease_generation,
    lease_owner_id,lease_heartbeat_at
FROM platform.jobs WHERE id=$2`, asyncACLAttemptID, asyncACLJobID); err != nil {
		t.Fatalf("runtime attempt insert through nested validator: %v", err)
	}
	if _, err := runtime.Exec(ctx, `
UPDATE platform.job_attempts
SET finished_at=clock_timestamp(),outcome='succeeded'
WHERE attempt_id=$1`, asyncACLAttemptID); err != nil {
		t.Fatalf("runtime attempt completion: %v", err)
	}
	if _, err := runtime.Exec(ctx, `
WITH stamp AS (
    SELECT GREATEST(clock_timestamp(),updated_at+interval '1 microsecond') AS at
    FROM platform.jobs WHERE id=$1
)
UPDATE platform.jobs SET state='succeeded',lease_owner_id=NULL,
    lease_heartbeat_at=NULL,lease_until=NULL,retain_until=stamp.at+interval '90 days',
    updated_at=stamp.at,completed_at=stamp.at
FROM stamp WHERE id=$1`, asyncACLJobID); err != nil {
		t.Fatalf("runtime job completion: %v", err)
	}
}

func assertAsyncACLDenials(
	t *testing.T, ctx context.Context, runtime *pgxpool.Conn,
) {
	t.Helper()
	for name, statement := range map[string]string{
		"immutable_job_update": `UPDATE platform.jobs SET kind='workflow.changed' WHERE id='` + asyncACLJobID + `'`,
		"job_delete":           `DELETE FROM platform.jobs WHERE id='` + asyncACLJobID + `'`,
		"inbox_update":         `UPDATE platform.inbox SET receipt_sha256=repeat('b',64) WHERE false`,
		"adoption_read":        `SELECT 1 FROM platform.async_v1_adoption_runs LIMIT 1`,
		"maintenance_execute":  `SELECT platform.platform_async_purge_jobs_v1(1)`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := runtime.Exec(ctx, statement)
			assertPermissionDenied(t, err)
		})
	}
}

func assertPermissionDenied(t *testing.T, err error) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("error = %v, want PostgreSQL 42501", err)
	}
}

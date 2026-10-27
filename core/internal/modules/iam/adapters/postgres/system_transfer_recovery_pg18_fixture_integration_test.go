//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

type iamExpiredReconciliationFixture struct {
	exportID      string
	originalJobID string
	originalKeyID string
	jobID         string
	attemptID     string
	leaseOwnerID  string
	effectID      string
	receiptID     string
	outcomeID     string
	generation    int64
	marker        string
}

var iamExpiredReconciliationFixtures = []iamExpiredReconciliationFixture{
	{
		exportID:      "019be100-0000-7000-8000-000000000001",
		originalJobID: "019be100-0000-7000-8000-000000000011",
		originalKeyID: "019be100-0000-7000-8000-000000000021",
		jobID:         "019be100-0000-7000-8000-000000000101",
		attemptID:     "019be100-0000-7000-8000-000000000201",
		leaseOwnerID:  "019be100-0000-7000-8000-000000000301",
		effectID:      "019be100-0000-7000-8000-000000000401",
		receiptID:     "019be100-0000-7000-8000-000000000501",
		outcomeID:     "019be100-0000-7000-8000-000000000601",
		generation:    1, marker: "41",
	},
	{
		exportID:      "019be100-0000-7000-8000-000000000002",
		originalJobID: "019be100-0000-7000-8000-000000000012",
		originalKeyID: "019be100-0000-7000-8000-000000000022",
		jobID:         "019be100-0000-7000-8000-000000000102",
		attemptID:     "019be100-0000-7000-8000-000000000202",
		leaseOwnerID:  "019be100-0000-7000-8000-000000000302",
		effectID:      "019be100-0000-7000-8000-000000000402",
		receiptID:     "019be100-0000-7000-8000-000000000502",
		outcomeID:     "019be100-0000-7000-8000-000000000602",
		generation:    2, marker: "42",
	},
}

const seedIAMExpiredReconciliationJob = `
INSERT INTO platform.jobs(
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 effect_idempotency_key,priority,run_at,state,attempts,max_attempts,
 lease_owner_id,lease_generation,lease_heartbeat_at,lease_until,
 effect_started_at,created_at,updated_at
) VALUES(
 $2,$1,'tenant',$1,'iam_system_resources',1,'iam',1,'v2',
 'iam.system-resource.export-reconcile',1,
 'iam-system-record/users/'||$3::text,
 encode(sha256(convert_to('iam-system-record/users/'||$3::text,'UTF8')),'hex'),
 'iam-system-transfer:iam.system-resource.export-reconcile:'||$3::text,
 'external','iam-system-export-reconcile:'||$3::text,90,
 clock_timestamp()-interval '5 minutes','running',$5::integer,8,$4,$5::integer,
 clock_timestamp()-interval '4 minutes',
 clock_timestamp()-interval '2 minutes',
 clock_timestamp()-interval '3 minutes',
 clock_timestamp()-interval '6 minutes',
 clock_timestamp()-interval '4 minutes 30 seconds'
)`

const seedIAMExpiredReconciliationAttempt = `
INSERT INTO platform.job_attempts(
 attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
 lease_owner_id,started_at
) SELECT $4,$2,'tenant',$1,$5::integer,$5::integer,$3,lease_heartbeat_at
  FROM platform.jobs WHERE id=$2`

const seedIAMExpiredReconciliationEffect = `
UPDATE platform.job_attempts AS attempt SET
 effect_started_at=job.effect_started_at,effect_id=$3,
 effect_request_sha256=repeat($4,32)
FROM platform.jobs AS job
WHERE attempt.attempt_id=$2 AND job.id=$1`

func seedIAMExpiredReconciliationPlatform(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	const privileges = `GRANT INSERT ON platform.jobs,platform.job_attempts
TO harpy_v2_runtime;
GRANT UPDATE ON platform.job_attempts TO harpy_v2_runtime;`
	if _, err := database.ExecContext(ctx, privileges); err != nil {
		t.Fatalf("grant IAM recovery fixture privileges: %v", err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve IAM recovery fixture connection: %v", err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx,
		"SET SESSION AUTHORIZATION harpy_v2_runtime"); err != nil {
		t.Fatalf("bind IAM recovery fixture session: %v", err)
	}
	for _, fixture := range iamExpiredReconciliationFixtures {
		_, err = connection.ExecContext(ctx, seedIAMExpiredReconciliationJob,
			iamFoundationTenants[0].id, fixture.jobID, fixture.exportID,
			fixture.leaseOwnerID, fixture.generation)
		if err != nil {
			t.Fatalf("seed reconciliation job %s: %v", fixture.jobID, err)
		}
		_, err = connection.ExecContext(ctx, seedIAMExpiredReconciliationAttempt,
			iamFoundationTenants[0].id, fixture.jobID, fixture.leaseOwnerID,
			fixture.attemptID, fixture.generation)
		if err != nil {
			t.Fatalf("seed reconciliation attempt %s: %v", fixture.attemptID, err)
		}
		_, err = connection.ExecContext(ctx, seedIAMExpiredReconciliationEffect,
			fixture.jobID, fixture.attemptID, fixture.effectID, fixture.marker)
		if err != nil {
			t.Fatalf("seed reconciliation effect %s: %v", fixture.effectID, err)
		}
	}
	if _, err = connection.ExecContext(ctx, "RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatalf("reset IAM recovery fixture session: %v", err)
	}
	const revoke = `REVOKE INSERT ON platform.jobs,platform.job_attempts
FROM harpy_v2_runtime;
REVOKE UPDATE ON platform.job_attempts FROM harpy_v2_runtime;`
	if _, err = database.ExecContext(ctx, revoke); err != nil {
		t.Fatalf("revoke IAM recovery fixture privileges: %v", err)
	}
}

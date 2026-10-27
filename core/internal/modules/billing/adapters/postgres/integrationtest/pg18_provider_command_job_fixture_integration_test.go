//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type providerPGRunner interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func createProviderPGCommand(
	t *testing.T,
	ctx context.Context,
	fixture *providerPGFixture,
) uuid.UUID {
	t.Helper()
	transaction, err := fixture.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	var commandID, finalizerID uuid.UUID
	var replay bool
	err = transaction.QueryRowContext(ctx, `
SELECT command_id,finalizer_job_id,replay
FROM platform.billing_create_provider_command_v1(
 $1,$2,$3,$4,$5,$6,$7,'change','provider-effect',$8,$9
)`, fixture.authorityID, fixture.tenantID, fixture.commandID,
		fixture.finalizerID, fixture.actorID, fixture.mappingID,
		fixture.subscriptionID, fixture.digest, fixture.now,
	).Scan(&commandID, &finalizerID, &replay)
	if err != nil || replay || commandID != fixture.commandID ||
		finalizerID != fixture.finalizerID {
		t.Fatalf("create provider command: ids=%s/%s replay=%t err=%v",
			commandID, finalizerID, replay, err)
	}
	insertProviderPGActiveJob(t, ctx, transaction, *fixture)
	attemptID := insertProviderPGFinalizer(t, ctx, transaction, *fixture,
		fixture.finalizerID)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return attemptID
}

func insertProviderPGActiveJob(
	t *testing.T, ctx context.Context, runner providerPGRunner, f providerPGFixture,
) {
	t.Helper()
	_, err := runner.ExecContext(ctx, `
INSERT INTO platform.jobs(
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 effect_idempotency_key,priority,run_at,state,attempts,max_attempts,
 lease_generation,created_at,updated_at
) VALUES($1,$2,'tenant',$2,'billing',1,'billing',1,'v2',
 'billing.provider.command',1,$3,$4,$5,'external',$5,100,$6,
 'queued',0,8,0,$6,$6)
`, f.commandID, f.tenantID,
		"billing-provider-command/testpay/"+f.commandID.String(), f.digestHex,
		"billing:provider:command:testpay:"+f.commandID.String(), f.now)
	if err != nil {
		t.Fatalf("insert provider active job: %v", err)
	}
}

func insertProviderPGFinalizer(
	t *testing.T, ctx context.Context, runner providerPGRunner,
	f providerPGFixture, jobID uuid.UUID,
) uuid.UUID {
	t.Helper()
	attemptID, leaseID := uuid.New(), uuid.New()
	dedupe := "billing:provider:finalize:" + f.commandID.String()
	if jobID != f.finalizerID {
		dedupe += ":" + jobID.String()
	}
	_, err := runner.ExecContext(ctx, `
INSERT INTO platform.jobs(
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 priority,run_at,state,attempts,max_attempts,lease_owner_id,lease_generation,
 lease_heartbeat_at,lease_until,created_at,updated_at
) VALUES($1,$2,'tenant',$2,'billing',1,'billing',1,'v2',
 'billing.provider.finalize',1,$3,$4,$5,'database_transactional',100,$6::timestamptz,
 'running',1,8,$7,1,$6::timestamptz,$6::timestamptz+interval '1 minute',
 $6::timestamptz,$6::timestamptz)
`, jobID, f.tenantID, "billing-provider-finalize/"+f.commandID.String(),
		f.digestHex, dedupe, f.now, leaseID)
	if err == nil {
		_, err = runner.ExecContext(ctx, `
INSERT INTO platform.job_attempts(
 attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
 lease_owner_id,started_at
) VALUES($1,$2,'tenant',$3,1,1,$4,$5)
`, attemptID, jobID, f.tenantID, leaseID, f.now)
	}
	if err != nil {
		t.Fatalf("insert provider finalizer: %v", err)
	}
	return attemptID
}

func makeProviderPGUnknown(
	t *testing.T,
	ctx context.Context,
	runner providerPGRunner,
	f providerPGFixture,
) {
	t.Helper()
	makeProviderPGJobUnknown(t, ctx, runner, f, f.commandID)
}

func makeProviderPGJobUnknown(
	t *testing.T,
	ctx context.Context,
	runner providerPGRunner,
	f providerPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	effectID, attemptID, leaseID := uuid.New(), uuid.New(), uuid.New()
	completed := f.now.Add(2 * time.Second)
	_, err := runner.ExecContext(ctx, `
UPDATE platform.jobs SET state='dead',attempts=1,lease_generation=1,
 effect_started_at=$2::timestamptz,last_error_code='JOB_OUTCOME_UNKNOWN',
 updated_at=$3::timestamptz,completed_at=$3::timestamptz,
 retain_until=$3::timestamptz+interval '366 days'
WHERE id=$1;
`, jobID, f.now.Add(time.Second), completed)
	if err == nil {
		_, err = runner.ExecContext(ctx, `
INSERT INTO platform.job_attempts(
 attempt_id,job_id,scope_kind,scope_id,attempt,lease_generation,
 lease_owner_id,started_at,finished_at,outcome,error_code,effect_started_at,
 effect_id,effect_request_sha256
) VALUES($1,$2,'tenant',$3,1,1,$4,$5,$6,'dead','JOB_OUTCOME_UNKNOWN',
 $7,$8,$9)
`, attemptID, jobID, f.tenantID, leaseID, f.now, completed,
			f.now.Add(time.Second), effectID, f.digestHex)
	}
	if err != nil {
		t.Fatalf("make provider outcome unknown: %v", err)
	}
}

func insertProviderPGReconciliation(
	t *testing.T,
	ctx context.Context,
	runner providerPGRunner,
	f providerPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	digest := sha256.Sum256([]byte("testpay\x00" + f.commandID.String()))
	key := "billing:provider:reconcile:" + fmt.Sprintf("%x", digest)
	_, err := runner.ExecContext(ctx, `
INSERT INTO platform.jobs(
 id,tenant_id,scope_kind,scope_id,authority_group_id,
 authority_group_generation,authority_domain,cutover_generation,writer_origin,
 kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
 effect_idempotency_key,priority,run_at,state,attempts,max_attempts,
 lease_generation,created_at,updated_at
) VALUES($1,$2,'tenant',$2,'billing',1,'billing',1,'v2',
 'billing.provider.reconcile',1,$3,$4,$5,'external',$5,100,$6,
 'queued',0,8,0,$6,$6)
`, jobID, f.tenantID,
		"billing-provider-reconciliation/testpay/"+f.commandID.String(),
		f.digestHex, key, f.now.Add(time.Second))
	if err != nil {
		t.Fatalf("insert provider reconciliation job: %v", err)
	}
}

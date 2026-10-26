//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func assertStaleDispositionCollisionRollback(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	tenantID := addStaleJobTenant(t, ctx, fixture)
	jobID := insertQueuedLifecycleJob(
		t, ctx, fixture, tenantID.String(), jobsv1.EffectDatabaseTransactional, 8,
	)
	setPublisherDomainGeneration(t, ctx, fixture.publisherPGFixture, tenantID, 2)
	if parked, err := fixture.store.ReconcileStaleJobs(ctx, 50); err != nil || parked != 1 {
		t.Fatalf("park collision candidate = %d, %v", parked, err)
	}
	authority, err := fixture.store.ResolveStaleJobAuthority(ctx, jobID)
	if err != nil {
		t.Fatal(err)
	}
	command := staleDispositionCommand(
		t, jobID, jobsv1.StaleJobAdopt, "adopt after collision rollback",
	)
	prepared, err := withJobAuthorityTransaction(
		ctx, fixture.store, authority,
		func(transaction pgx.Tx) (preparedStaleJobDisposition, error) {
			source, inspectErr := inspectStaleJobDisposition(
				ctx, transaction, authority, jobID,
			)
			if inspectErr != nil {
				return preparedStaleJobDisposition{}, inspectErr
			}
			return source.prepare(authority, command)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	insertStaleDispositionCollision(t, ctx, fixture, tenantID.String(), jobID,
		prepared.resultJobID.String())
	if _, err := fixture.store.DisposeStaleJob(
		ctx, authority, command,
	); !errors.Is(err, ErrStaleJobDispositionConflict) {
		t.Fatalf("replacement collision error = %v", err)
	}
	var state string
	var parkedAt any
	var dispositionRoots int
	err = fixture.owner.QueryRowContext(ctx, `
SELECT job.state,job.parked_at,
 (SELECT count(*) FROM platform.dead_letters AS root
  WHERE root.source_kind='job' AND root.source_id=job.id
    AND root.redacted_envelope->>'contract_version'='stale-job-disposition.v1')
FROM platform.jobs AS job WHERE job.id=$1`, jobID).Scan(
		&state, &parkedAt, &dispositionRoots,
	)
	if err != nil || state != "queued" || parkedAt == nil || dispositionRoots != 0 {
		t.Fatalf("collision rollback = %s/%v roots=%d: %v",
			state, parkedAt, dispositionRoots, err)
	}
	if _, err := fixture.owner.ExecContext(
		ctx, "DELETE FROM platform.jobs WHERE id=$1", prepared.resultJobID,
	); err != nil {
		t.Fatalf("remove replacement collision: %v", err)
	}
	result, err := fixture.store.DisposeStaleJob(ctx, authority, command)
	if err != nil || !result.Created() {
		t.Fatalf("adoption after collision rollback = %#v, %v", result, err)
	}
}

func insertStaleDispositionCollision(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	tenantID string,
	sourceID any,
	resultID string,
) {
	t.Helper()
	_, err := fixture.runtime.Exec(ctx, `
INSERT INTO platform.jobs (
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,writer_origin,
    kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
    priority,run_at,state,attempts,max_attempts,lease_generation,retry_of_job_id
) VALUES (
    $1,$2,'tenant',$2,'operations',1,'workflow',2,'v2',
    'workflow.execute',1,$3,repeat('b',64),$4,'database_transactional',10,
    pg_catalog.clock_timestamp(),'queued',0,8,0,$5
)`, resultID, tenantID, "payload/collision/"+resultID,
		"collision:"+resultID, sourceID)
	if err != nil {
		t.Fatalf("insert replacement collision: %v", err)
	}
}

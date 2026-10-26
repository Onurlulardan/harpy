//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type jobLifecyclePGFixture struct {
	publisherPGFixture
	authority JobAuthority
}

func newJobLifecyclePGFixture(
	t *testing.T,
	ctx context.Context,
	origin string,
) jobLifecyclePGFixture {
	t.Helper()
	base := newPublisherPGFixture(t, ctx, origin)
	grantJobLifecycleRuntime(t, ctx, base.owner)
	authority, err := newJobAuthority(
		base.store.identity, "tenant", uuid.MustParse(asyncTenantA),
		"operations", 1, "workflow", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	return jobLifecyclePGFixture{publisherPGFixture: base, authority: authority}
}

func grantJobLifecycleRuntime(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
CREATE TABLE platform.job_lifecycle_probe (
    tenant_id uuid NOT NULL, marker text NOT NULL,
    PRIMARY KEY (tenant_id, marker)
);
GRANT SELECT, INSERT ON platform.job_lifecycle_probe TO harpy_runtime;
GRANT INSERT (
    lease_owner_id,lease_heartbeat_at,lease_until,
    effect_started_at,created_at,updated_at
) ON platform.jobs TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("grant job lifecycle runtime: %v", err)
	}
}

func lifecycleCatalog(
	t *testing.T,
	effect jobsv1.EffectClass,
	maximumAttempts int,
) jobsv1.Catalog {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(
		"workflow.execute", 1, "workflow", effect, maximumAttempts,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func insertQueuedLifecycleJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	tenantID string,
	effect jobsv1.EffectClass,
	maximumAttempts int,
) uuid.UUID {
	t.Helper()
	jobID := uuid.New()
	var effectKey any
	if effect == jobsv1.EffectExternal {
		effectKey = "effect:workflow:" + jobID.String()
	}
	const statement = `
INSERT INTO platform.jobs (
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,writer_origin,
    kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
    effect_idempotency_key,priority,run_at,state,attempts,max_attempts,lease_generation
) VALUES (
    $1,$2,'tenant',$2,'operations',1,'workflow',1,'v2',
    'workflow.execute',1,$3,repeat('a',64),$4,$5,$6,10,
    pg_catalog.clock_timestamp()-interval '1 second','queued',0,$7,0
)`
	_, err := fixture.runtime.Exec(
		ctx, statement, jobID, tenantID, "payload/"+jobID.String(),
		"dedupe:"+jobID.String(), string(effect), effectKey, maximumAttempts,
	)
	if err != nil {
		t.Fatalf("insert queued lifecycle job: %v", err)
	}
	return jobID
}

type lifecycleBoundPGTx struct{ transaction pgx.Tx }

func (database lifecycleBoundPGTx) Execute(
	ctx context.Context,
	statement string,
	arguments ...any,
) (platformpostgres.CommandResult, error) {
	_, err := database.transaction.Exec(ctx, statement, arguments...)
	return platformpostgres.CommandResult{}, err
}

func (database lifecycleBoundPGTx) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) platformpostgres.Row {
	return database.transaction.QueryRow(ctx, statement, arguments...)
}

func lifecycleAuthorityStamp(authority JobAuthority) authorityStamp {
	return authorityStamp{
		scopeID: authority.scopeID, groupID: authority.groupID,
		groupGeneration: authority.groupGeneration, domain: authority.domain,
		generation: authority.generation,
	}
}

func lifecycleState(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	jobID uuid.UUID,
) (string, int) {
	t.Helper()
	var state string
	var openAttempts int
	err := fixture.owner.QueryRowContext(ctx, `
SELECT job.state, pg_catalog.count(*) FILTER (WHERE attempt.finished_at IS NULL)
FROM platform.jobs AS job
LEFT JOIN platform.job_attempts AS attempt ON attempt.job_id=job.id
WHERE job.id=$1 GROUP BY job.state`, jobID).Scan(&state, &openAttempts)
	if err != nil {
		t.Fatal(fmt.Errorf("read lifecycle state: %w", err))
	}
	return state, openAttempts
}

func lifecycleReceiptTime(startedAt time.Time) string {
	return startedAt.Add(time.Second).UTC().Format(time.RFC3339Nano)
}

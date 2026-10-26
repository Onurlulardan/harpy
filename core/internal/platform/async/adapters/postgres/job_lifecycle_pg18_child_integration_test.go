//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func assertCrossKindChildJobAtomicity(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	parentID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantA, jobsv1.EffectDatabaseTransactional, 8,
	)
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != parentID {
		t.Fatalf("parent child-job claim = %#v, %v", claims, err)
	}
	claim := claims[0]
	policy, catalog := pgChildJobCatalog(t)
	child := pgChildJob(claim, policy)

	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(context.Background())
	bound := &boundDatabaseJobStore{
		database: lifecycleBoundPGTx{transaction: transaction}, jobs: catalog,
	}
	if err := bound.assertAuthority(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if err := bound.enqueueChild(ctx, claim, child); err != nil {
		t.Fatal(err)
	}
	conflict := child
	conflict.PayloadSHA256 = strings.Repeat("e", 64)
	if err := bound.enqueueChild(ctx, claim, conflict); !errors.Is(
		err, ErrIdempotencyConflict,
	) {
		t.Fatalf("child idempotency conflict = %v", err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if pgChildJobCount(t, ctx, fixture, child.JobID) != 0 {
		t.Fatal("rolled-back child job remained visible")
	}

	transaction, err = fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(context.Background())
	bound = &boundDatabaseJobStore{
		database: lifecycleBoundPGTx{transaction: transaction}, jobs: catalog,
	}
	if err := bound.assertAuthority(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if err := bound.enqueueChild(ctx, claim, child); err != nil {
		t.Fatal(err)
	}
	if err := bound.enqueueChild(ctx, claim, child); err != nil {
		t.Fatalf("exact child replay = %v", err)
	}
	if err := bound.complete(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertPersistedChildJob(t, ctx, fixture, claim, child, policy)
}

func pgChildJobCatalog(t *testing.T) (jobsv1.KindPolicy, jobsv1.Catalog) {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(
		"workflow.reconcile", 1, "workflow", jobsv1.EffectExternal, 6,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return policy, catalog
}

func pgChildJob(
	claim JobClaim,
	policy jobsv1.KindPolicy,
) platformasync.DatabaseJobChildInput {
	return platformasync.DatabaseJobChildInput{
		JobID: uuid.Must(uuid.NewV7()), Kind: policy.Kind,
		PayloadRevision: policy.PayloadRevision,
		RunAt: claim.StartedAt().UTC().Truncate(time.Microsecond).
			Add(time.Microsecond),
		PayloadReference:     "workflow-reconciliation/" + claim.JobID().String(),
		PayloadSHA256:        strings.Repeat("d", 64),
		DedupeKey:            "workflow:reconcile:" + claim.JobID().String(),
		EffectIdempotencyKey: "workflow:effect:" + claim.JobID().String(),
		Priority:             80,
	}
}

func pgChildJobCount(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	jobID uuid.UUID,
) int {
	t.Helper()
	var count int
	if err := fixture.owner.QueryRowContext(
		ctx, "SELECT count(*) FROM platform.jobs WHERE id=$1", jobID,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func assertPersistedChildJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	claim JobClaim,
	child platformasync.DatabaseJobChildInput,
	policy jobsv1.KindPolicy,
) {
	t.Helper()
	var tenantID uuid.UUID
	var groupID, domain, kind, effectClass, effectKey, state string
	var groupGeneration, generation int64
	var attempts int
	err := fixture.owner.QueryRowContext(ctx, `
SELECT tenant_id,authority_group_id,authority_group_generation,
 authority_domain,cutover_generation,kind,effect_class,
 effect_idempotency_key,state,max_attempts
FROM platform.jobs WHERE id=$1`, child.JobID).Scan(
		&tenantID, &groupID, &groupGeneration, &domain, &generation,
		&kind, &effectClass, &effectKey, &state, &attempts,
	)
	parentTenant, _ := claim.TenantID()
	if err != nil || tenantID != parentTenant || groupID != claim.authority.groupID ||
		groupGeneration != claim.authority.groupGeneration ||
		domain != claim.authority.domain || generation != claim.authority.generation ||
		kind != child.Kind || effectClass != string(policy.EffectClass) ||
		effectKey != child.EffectIdempotencyKey || state != "queued" ||
		attempts != policy.MaximumAttempts {
		t.Fatalf("persisted child mismatch: %v %s/%s/%s", err, kind, effectClass, state)
	}
}

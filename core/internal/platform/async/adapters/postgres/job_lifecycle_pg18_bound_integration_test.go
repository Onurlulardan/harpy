//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func assertBoundDatabaseJobAtomicity(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	jobID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantA, jobsv1.EffectDatabaseTransactional, 8,
	)
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != jobID {
		t.Fatalf("database job claim = %#v, %v", claims, err)
	}
	claim := claims[0]
	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	firstTransaction := transaction
	defer firstTransaction.Rollback(context.Background())
	bound := &repository{
		database:  lifecycleBoundPGTx{transaction: transaction},
		authority: lifecycleAuthorityStamp(fixture.authority),
	}
	if _, err := transaction.Exec(
		ctx, "INSERT INTO platform.job_lifecycle_probe VALUES($1,$2)",
		asyncTenantA, "rolled-back",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := bound.CompleteDatabaseJob(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	state, open := lifecycleState(t, ctx, fixture, jobID)
	if state != string(jobsv1.StateRunning) || open != 1 ||
		lifecycleProbeCount(t, ctx, fixture, "rolled-back") != 0 {
		t.Fatalf("rollback state=%s open=%d", state, open)
	}

	transaction, err = fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	secondTransaction := transaction
	defer secondTransaction.Rollback(context.Background())
	bound.database = lifecycleBoundPGTx{transaction: transaction}
	if _, err := transaction.Exec(
		ctx, "INSERT INTO platform.job_lifecycle_probe VALUES($1,$2)",
		asyncTenantA, "committed",
	); err != nil {
		t.Fatal(err)
	}
	if _, err := bound.CompleteDatabaseJob(ctx, claim); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	state, open = lifecycleState(t, ctx, fixture, jobID)
	if state != string(jobsv1.StateSucceeded) || open != 0 ||
		lifecycleProbeCount(t, ctx, fixture, "committed") != 1 {
		t.Fatalf("commit state=%s open=%d", state, open)
	}
	assertCrossTenantBoundJobRejected(t, ctx, fixture)
}

func assertCrossTenantBoundJobRejected(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	claim.authority = fixture.authority
	claim.authority.owner = fixture.store.identity
	claim.tenantID = uuid.MustParse(asyncTenantA)
	claim.leaseOwnerID = uuid.New()
	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback(ctx)
	other := lifecycleAuthorityStamp(fixture.authority)
	other.scopeID = uuid.MustParse(asyncTenantB)
	bound := &repository{
		database: lifecycleBoundPGTx{transaction: transaction}, authority: other,
	}
	if _, err := bound.CompleteDatabaseJob(
		ctx, claim,
	); !errors.Is(err, ErrAuthorityMismatch) {
		t.Fatalf("cross-tenant bound completion error = %v", err)
	}
}

func lifecycleProbeCount(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	marker string,
) int {
	t.Helper()
	var count int
	if err := fixture.owner.QueryRowContext(
		ctx, "SELECT count(*) FROM platform.job_lifecycle_probe WHERE marker=$1", marker,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

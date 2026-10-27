//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func assertTerminalJobDeadLetterAtomicity(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	jobID, claim := claimLifecycleDatabaseJob(t, ctx, fixture, 8)
	failure := terminalJobFailure(
		t, jobsv1.ErrorPermanent, "WORKFLOW_REJECTED", "e",
	)
	transaction, bound := beginBoundLifecycle(t, ctx, fixture)
	result, err := bound.FailJobBeforeEffect(ctx, claim, failure)
	if err != nil || result.State() != jobsv1.StateDead {
		t.Fatalf("stage permanent terminal failure=%#v %v", result, err)
	}
	var stagedRoots int
	if err := transaction.QueryRow(ctx, `
SELECT count(*) FROM platform.dead_letters WHERE source_kind='job' AND source_id=$1`,
		jobID).Scan(&stagedRoots); err != nil || stagedRoots != 1 {
		t.Fatalf("staged terminal roots=%d %v", stagedRoots, err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	state, open := lifecycleState(t, ctx, fixture, jobID)
	if state != string(jobsv1.StateRunning) || open != 1 ||
		terminalJobRootCount(t, ctx, fixture, jobID) != 0 {
		t.Fatalf("terminal rollback state=%s open=%d", state, open)
	}
	if err := commitBoundFailure(ctx, fixture, claim, failure); err != nil {
		t.Fatal(err)
	}
	assertTerminalJobRoot(t, ctx, fixture, jobID, "permanent", "WORKFLOW_REJECTED")
	if err := commitBoundFailure(ctx, fixture, claim, failure); !errors.Is(err, ErrJobLeaseLost) {
		t.Fatalf("replayed terminal claim error=%v", err)
	}
	if terminalJobRootCount(t, ctx, fixture, jobID) != 1 {
		t.Fatal("replayed terminal claim duplicated its root")
	}
}

func assertTerminalJobDeadLetterConcurrency(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	jobID, claim := claimLifecycleDatabaseJob(t, ctx, fixture, 1)
	failure := terminalJobFailure(
		t, jobsv1.ErrorTransient, "WORKFLOW_RETRY_LIMIT", "f",
	)
	errorsSeen := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errorsSeen <- commitBoundFailure(ctx, fixture, claim, failure)
		}()
	}
	wait.Wait()
	close(errorsSeen)
	succeeded, fenced := 0, 0
	for err := range errorsSeen {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, ErrJobLeaseLost):
			fenced++
		default:
			t.Fatalf("concurrent terminal failure error=%v", err)
		}
	}
	if succeeded != 1 || fenced != 1 || terminalJobRootCount(t, ctx, fixture, jobID) != 1 {
		t.Fatalf("terminal race succeeded=%d fenced=%d", succeeded, fenced)
	}
	assertTerminalJobRoot(
		t, ctx, fixture, jobID, "delivery_limit", "WORKFLOW_RETRY_LIMIT",
	)
	if result, err := fixture.store.RecoverExpiredJobs(ctx, fixture.authority, 50); err != nil || result.Total() != 0 || terminalJobRootCount(t, ctx, fixture, jobID) != 1 {
		t.Fatalf("terminal restart recovery=%#v error=%v", result, err)
	}
}

func claimLifecycleDatabaseJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	maximumAttempts int,
) (uuid.UUID, JobClaim) {
	t.Helper()
	jobID := insertQueuedLifecycleJob(
		t, ctx, fixture, asyncTenantA,
		jobsv1.EffectDatabaseTransactional, maximumAttempts,
	)
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, maximumAttempts),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != jobID {
		t.Fatalf("claim terminal job=%#v %v", claims, err)
	}
	return jobID, claims[0]
}

func terminalJobFailure(
	t *testing.T,
	class jobsv1.ErrorClass,
	code string,
	digestByte string,
) JobFailure {
	t.Helper()
	failure, err := NewJobFailure(class, code, strings.Repeat(digestByte, 64), 0)
	if err != nil {
		t.Fatal(err)
	}
	return failure
}

func beginBoundLifecycle(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) (pgx.Tx, *repository) {
	t.Helper()
	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return transaction, &repository{
		database:  lifecycleBoundPGTx{transaction: transaction},
		authority: lifecycleAuthorityStamp(fixture.authority),
	}
}

func commitBoundFailure(
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	claim JobClaim,
	failure JobFailure,
) error {
	transaction, err := fixture.runtime.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer transaction.Rollback(context.Background())
	bound := &repository{
		database:  lifecycleBoundPGTx{transaction: transaction},
		authority: lifecycleAuthorityStamp(fixture.authority),
	}
	if _, err := bound.FailJobBeforeEffect(ctx, claim, failure); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

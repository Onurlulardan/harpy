//go:build integration

package subscription

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

type actorRaceResult struct {
	replay bool
	err    error
}

func TestSubscriptionActorCancellationRaceIsExactlyOnceOnPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startSubscriptionPG18(t, ctx)
	fixture := seedSubscriptionActor(t, ctx, database.runtime)
	sourceID := uuid.Must(uuid.NewV7())
	start := make(chan struct{})
	results := make(chan actorRaceResult, 2)
	var ready sync.WaitGroup
	ready.Add(2)
	for range 2 {
		go func() {
			ready.Done()
			<-start
			result, err := runActorCancellation(
				ctx, database.runtime, fixture, sourceID, "customer request", 1,
			)
			results <- actorRaceResult{replay: result.Replay, err: err}
		}()
	}
	ready.Wait()
	close(start)
	first, second := <-results, <-results
	if first.err != nil || second.err != nil || first.replay == second.replay {
		t.Fatalf(
			"race errors = %v / %v; replays = %t / %t",
			first.err, second.err, first.replay, second.replay,
		)
	}
	assertActorEffectCounts(t, ctx, database.owner, fixture, 1, 1, 1, 1)
	if _, err := runActorCancellation(
		ctx, database.runtime, fixture, uuid.Must(uuid.NewV7()),
		"stale request", 1,
	); !errors.Is(err, subscriptionapp.ErrStaleRevision) {
		t.Fatalf("stale revision error = %v", err)
	}
	if _, err := runActorCancellation(
		ctx, database.runtime, fixture, sourceID,
		"different digest", 1,
	); !errors.Is(err, subscriptionapp.ErrIdempotencyReused) {
		t.Fatalf("reused source error = %v", err)
	}
	assertActorEffectCounts(t, ctx, database.owner, fixture, 1, 1, 1, 1)
}

func runActorCancellation(
	ctx context.Context,
	database *sql.DB,
	fixture subscriptionActorFixture,
	sourceID uuid.UUID,
	reason string,
	expectedRevision int64,
) (subscriptionapp.Result, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL ROLE harpy_owner`); err != nil {
		return subscriptionapp.Result{}, err
	}
	storage := &store{
		tenantID: fixture.tenantID, actorID: fixture.actorID, mode: modeActor,
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return transaction.QueryRowContext(ctx, statement, arguments...)
		},
		execute: func(ctx context.Context, statement string, arguments ...any) error {
			_, executeErr := transaction.ExecContext(ctx, statement, arguments...)
			return executeErr
		},
		allowed:  func(context.Context) bool { return true },
		notFound: func(err error) bool { return errors.Is(err, sql.ErrNoRows) },
		effects: &subscriptionActorEffects{
			transaction: transaction, tenantID: fixture.tenantID,
		},
	}
	commands, err := subscriptionapp.NewCommands(
		storage, uuid.NewV7, func() time.Time { return fixture.now },
	)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	result, err := commands.CancelAtPeriodEnd(ctx, subscriptionapp.CancelAtPeriodEndRequest{
		ActorID: fixture.actorID, SubscriptionID: fixture.subscriptionID,
		SourceOperationID: sourceID, ExpectedRevision: expectedRevision, Reason: reason,
	})
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	if err := transaction.Commit(); err != nil {
		return subscriptionapp.Result{}, err
	}
	return result, nil
}

func assertActorEffectCounts(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture subscriptionActorFixture,
	receipts, audits, jobs, events int,
) {
	t.Helper()
	queries := []struct {
		statement string
		want      int
		scoped    bool
	}{
		{`SELECT count(*) FROM platform.billing_subscription_operations WHERE subscription_id=$1`, receipts, true},
		{`SELECT count(*) FROM platform.billing_subscription_audit WHERE subscription_id=$1`, audits, true},
		{`SELECT count(*) FROM platform.billing_subscription_boundary_jobs WHERE subscription_id=$1`, jobs, true},
		{`SELECT count(*) FROM platform.test_subscription_events`, events, false},
	}
	for _, query := range queries {
		var got int
		arguments := []any(nil)
		if query.scoped {
			arguments = append(arguments, fixture.subscriptionID)
		}
		if err := database.QueryRowContext(ctx, query.statement, arguments...).Scan(&got); err != nil || got != query.want {
			t.Fatalf("effect count for %q = %d, %v; want %d", query.statement, got, err, query.want)
		}
	}
}

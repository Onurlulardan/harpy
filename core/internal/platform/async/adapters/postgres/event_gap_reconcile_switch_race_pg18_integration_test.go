//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

func TestEventGapReconcileAuthoritySwitchRaceOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertEventGapReconcileAuthoritySwitchRace(t, ctx, &fixture)
		})
	}
}

func assertEventGapReconcileAuthoritySwitchRace(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	for sequence := int64(1); sequence <= 3; sequence++ {
		fixture.seedDelivered(t, ctx, sequence)
	}
	if fixture.deliver(t, ctx, 3) != brokerrabbit.HandleCommitted {
		t.Fatal("switch-race gap was not parked")
	}
	fixture.deliver(t, ctx, 1)
	fixture.deliver(t, ctx, 2)
	_, jobID := fixture.gapIDs(t, ctx, 3)
	claim := fixture.claimGap(t, ctx, jobID)
	eventID := uuid.MustParse(fixture.envelopes[3].EventID)
	before := readSourceStaleOutbox(t, ctx, fixture, eventID)
	switchTransaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer switchTransaction.Rollback()
	if _, err := switchTransaction.ExecContext(ctx, `
UPDATE platform.cutover_groups SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='operations'`,
		fixture.tenantID); err != nil {
		t.Fatal(err)
	}
	type reconcileOutcome struct {
		result EventGapReconciliation
		err    error
	}
	done := make(chan reconcileOutcome, 1)
	go func() {
		result, reconcileErr := fixture.store.ReconcileEventGap(ctx, claim)
		done <- reconcileOutcome{result: result, err: reconcileErr}
	}()
	select {
	case outcome := <-done:
		t.Fatalf("reconcile escaped open producer switch: %#v/%v",
			outcome.result, outcome.err)
	case <-time.After(200 * time.Millisecond):
	}
	if err := switchTransaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var outcome reconcileOutcome
	select {
	case outcome = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("reconcile remained blocked after producer switch commit")
	}
	if outcome.err != nil || outcome.result.Requeued() ||
		outcome.result.Disposition() != EventGapSourceAuthorityStale {
		t.Fatalf("switch-race reconciliation = %#v/%v", outcome.result, outcome.err)
	}
	if after := readSourceStaleOutbox(t, ctx, fixture, eventID); after != before {
		t.Fatalf("switch-race source outbox mutated: before=%+v after=%+v", before, after)
	}
	assertSourceStaleGapTerminal(t, ctx, fixture, jobID, eventID)
	if _, err := fixture.store.ReconcileEventGap(
		ctx, claim,
	); !errors.Is(err, ErrJobLeaseLost) {
		t.Fatalf("switch-race replay error = %v", err)
	}
	var actions int
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FROM platform.dead_letters
WHERE id=$1 AND record_kind='repair_action'`, jobID).Scan(&actions); err != nil {
		t.Fatal(err)
	}
	if actions != 1 {
		t.Fatalf("switch-race terminal actions = %d, want 1", actions)
	}
}

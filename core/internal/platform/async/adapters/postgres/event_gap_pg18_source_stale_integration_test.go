//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

func TestEventGapSourceAuthorityStaleOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertEventGapSourceAuthorityStale(t, ctx, &fixture)
		})
	}
}

type sourceStaleOutboxSnapshot struct {
	state                             string
	nextAttempt, confirmed, delivered time.Time
	confirmSHA                        string
	updatedAt                         time.Time
}

func assertEventGapSourceAuthorityStale(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	fixture.seedDelivered(t, ctx, 1)
	fixture.seedDelivered(t, ctx, 2)
	if fixture.deliver(t, ctx, 2) != brokerrabbit.HandleCommitted {
		t.Fatal("source-stale gap was not durably parked")
	}
	_, jobID := fixture.gapIDs(t, ctx, 2)
	if fixture.gapReady(t, ctx, jobID) {
		t.Fatal("gap was ready before source authority became stale")
	}
	eventID := uuid.MustParse(fixture.envelopes[2].EventID)
	before := readSourceStaleOutbox(t, ctx, fixture, eventID)
	advanceEventSourceGroupOnly(t, ctx, fixture, 2)
	if !fixture.gapReady(t, ctx, jobID) {
		t.Fatal("source-stale gap did not become terminally claimable")
	}
	authority, err := newJobAuthority(
		fixture.store.identity, "tenant", fixture.tenantID,
		"workflow", 1, "workflow", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := fixture.store.ClaimJobs(
		ctx, authority, uuid.New(), 1,
		consumerJobCatalogForPipeline(fixture.consumers),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != jobID {
		t.Fatalf("claim source-stale event gap = %#v, %v", claims, err)
	}
	claim := claims[0]
	result, err := fixture.store.ReconcileEventGap(ctx, claim)
	if err != nil {
		t.Fatalf("reconcile source-stale gap: %v", err)
	}
	if result.Requeued() || result.ActionID() != jobID ||
		result.Disposition() != EventGapSourceAuthorityStale {
		t.Fatalf("source-stale result = %#v", result)
	}
	after := readSourceStaleOutbox(t, ctx, fixture, eventID)
	if before != after {
		t.Fatalf("source outbox mutated: before=%+v after=%+v", before, after)
	}
	assertSourceStaleGapTerminal(t, ctx, fixture, jobID, eventID)
	claims, err = fixture.store.ClaimJobs(
		ctx, authority, uuid.New(), 1,
		consumerJobCatalogForPipeline(fixture.consumers),
	)
	if err != nil || len(claims) != 0 {
		t.Fatalf("terminal source-stale gap reclaimed = %#v, %v", claims, err)
	}
}

func advanceEventSourceGroupOnly(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	generation int64,
) {
	t.Helper()
	result, err := fixture.owner.ExecContext(ctx, `
UPDATE platform.cutover_groups SET generation=$1
WHERE scope_kind='tenant' AND scope_id=$2 AND group_id='operations'`,
		generation, fixture.tenantID)
	if err != nil {
		t.Fatal(err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		t.Fatalf("advance source group affected %d: %v", affected, err)
	}
}

func readSourceStaleOutbox(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	eventID uuid.UUID,
) sourceStaleOutboxSnapshot {
	t.Helper()
	var value sourceStaleOutboxSnapshot
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT delivery_state,next_attempt_at,confirmed_at,delivered_at,
       confirm_sha256,updated_at FROM platform.outbox WHERE event_id=$1`,
		eventID).Scan(&value.state, &value.nextAttempt, &value.confirmed,
		&value.delivered, &value.confirmSHA, &value.updatedAt); err != nil {
		t.Fatal(err)
	}
	return value
}

func assertSourceStaleGapTerminal(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
	eventID uuid.UUID,
) {
	t.Helper()
	var state, jobCode, outcome, attemptCode, actionKind, reason string
	err := fixture.owner.QueryRowContext(ctx, `
SELECT job.state,job.last_error_code,attempt.outcome,attempt.error_code,
       action.record_kind,action.action_reason
FROM platform.jobs AS job
JOIN platform.job_attempts AS attempt ON attempt.job_id=job.id
JOIN platform.dead_letters AS action ON action.id=job.id
JOIN platform.dead_letters AS root ON root.id=action.root_id
WHERE job.id=$1 AND root.source_id=$2`, jobID, eventID).Scan(
		&state, &jobCode, &outcome, &attemptCode, &actionKind, &reason,
	)
	if err != nil {
		t.Fatal(err)
	}
	if state != "cancelled" || jobCode != "EVENT_GAP_SOURCE_AUTHORITY_STALE" ||
		outcome != "cancelled" || attemptCode != jobCode ||
		actionKind != "repair_action" || reason !=
		"automatic sequence-gap reconciliation terminated because source authority is stale" {
		t.Fatalf("source-stale terminal = %s/%s %s/%s %s/%s",
			state, jobCode, outcome, attemptCode, actionKind, reason)
	}
}

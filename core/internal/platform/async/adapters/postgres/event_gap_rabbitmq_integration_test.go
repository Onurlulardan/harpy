//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

func TestRabbitEventGapReplaysSameEventWithSingleEffect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	rabbit := startPipelineRabbit(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			fixture := newEventGapPGFixtureWithRabbit(t, ctx, origin, rabbit.url)
			assertRabbitEventGapReplay(t, ctx, &fixture)
		})
	}
}

func assertRabbitEventGapReplay(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	stop := startPipelineConsumer(
		t, ctx, fixture.worker.rabbit, fixture.catalog,
		fixture.consumers, fixture.handler,
	)
	defer stop()
	publisher, connection := openPipelinePublisher(
		t, ctx, fixture.worker.rabbit, fixture.consumers,
	)
	defer connection.Close()
	defer publisher.Close()
	fixture.seedDelivered(t, ctx, 1)
	fixture.seedDelivered(t, ctx, 2)

	publishGapEnvelope(t, ctx, publisher, fixture.envelopes[2], 1)
	fixture.handler.requireOutcome(t, ctx, brokerrabbit.HandleCommitted)
	rootID, jobID := fixture.gapIDs(t, ctx, 2)
	assertSingleGapEvidence(t, ctx, fixture, rootID, jobID)

	publishGapEnvelope(t, ctx, publisher, fixture.envelopes[1], 1)
	fixture.handler.requireOutcome(t, ctx, brokerrabbit.HandleCommitted)
	if !fixture.gapReady(t, ctx, jobID) {
		t.Fatal("Rabbit predecessor commit did not release gap job")
	}
	claim := fixture.claimGap(t, ctx, jobID)
	result, err := fixture.store.ReconcileEventGap(ctx, claim)
	if err != nil || !result.Requeued() || result.ActionID() != jobID {
		t.Fatalf("Rabbit gap reconcile = %#v, %v", result, err)
	}

	authority := publisherAuthorityFor(
		t, ctx, fixture.publisherPGFixture, fixture.tenantID,
	)
	replay := claimPipelineEvent(
		t, ctx, fixture.publisherPGFixture, authority,
		uuid.MustParse(fixture.envelopes[2].EventID), 2,
	)
	receipt, err := publisher.Publish(ctx, brokerrabbit.PublishRequest{
		Envelope: replay.Envelope(), LeaseGeneration: replay.LeaseGeneration(),
	})
	if err != nil {
		t.Fatalf("publish reconciled gap: %v", err)
	}
	if _, err := fixture.store.ConfirmPublish(
		ctx, replay, receipt.ConfirmSHA256,
	); err != nil {
		t.Fatalf("confirm reconciled gap: %v", err)
	}
	fixture.handler.requireOutcome(t, ctx, brokerrabbit.HandleCommitted)

	publishGapEnvelope(t, ctx, publisher, fixture.envelopes[2], 3)
	fixture.handler.requireOutcome(t, ctx, brokerrabbit.HandleDuplicate)
	assertGapEffects(t, ctx, fixture, 2)
	assertRabbitGapFinalState(t, ctx, fixture, jobID)
}

func publishGapEnvelope(
	t *testing.T,
	ctx context.Context,
	publisher *brokerrabbit.Publisher,
	envelope eventsv1.Envelope,
	leaseGeneration int64,
) {
	t.Helper()
	if _, err := publisher.Publish(ctx, brokerrabbit.PublishRequest{
		Envelope: envelope, LeaseGeneration: leaseGeneration,
	}); err != nil {
		t.Fatalf("publish event %s to RabbitMQ: %v", envelope.EventID, err)
	}
}

func assertRabbitGapFinalState(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
) {
	t.Helper()
	var attempts, actions int
	var state string
	err := fixture.owner.QueryRowContext(ctx, `
SELECT event.delivery_attempts,event.delivery_state,
 (SELECT count(*) FROM platform.dead_letters
  WHERE id=$2 AND record_kind='retry_action' AND source_id=event.event_id)
FROM platform.outbox AS event WHERE event.event_id=$1`,
		fixture.envelopes[2].EventID, jobID,
	).Scan(&attempts, &state, &actions)
	if err != nil || attempts != 2 || state != "delivered" || actions != 1 {
		t.Fatalf("Rabbit gap final attempts/state/actions = %d/%s/%d: %v",
			attempts, state, actions, err)
	}
}

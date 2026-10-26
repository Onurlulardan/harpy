//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

type eventGapPGFixture struct {
	publisherPGFixture
	worker    pipelineWorkerRuntime
	consumers eventsv1.ConsumerCatalog
	handler   *pipelineConsumerHandler
	tenantID  uuid.UUID
	aggregate uuid.UUID
	envelopes map[int64]eventsv1.Envelope
}

func newEventGapPGFixture(
	t *testing.T,
	ctx context.Context,
	origin string,
) eventGapPGFixture {
	return newEventGapPGFixtureWithRabbit(
		t, ctx, origin, "amqp://unused:unused@127.0.0.1:5672/unused",
	)
}

func newEventGapPGFixtureWithRabbit(
	t *testing.T,
	ctx context.Context,
	origin string,
	rabbitURL string,
) eventGapPGFixture {
	t.Helper()
	base := newPublisherPGFixture(t, ctx, origin)
	worker := preparePipelineRuntime(t, ctx, base, rabbitURL)
	_, consumers := consumerCatalogs(t, eventsv1.EffectDatabaseTransactional)
	tenantID := uuid.MustParse(asyncTenantA)
	return eventGapPGFixture{
		publisherPGFixture: base, worker: worker, consumers: consumers,
		handler:  newPipelineConsumerHandler(t, worker.coordinator, base.catalog, consumers),
		tenantID: tenantID, aggregate: uuid.New(),
		envelopes: make(map[int64]eventsv1.Envelope),
	}
}

func (fixture *eventGapPGFixture) seedDelivered(
	t *testing.T,
	ctx context.Context,
	sequence int64,
) eventsv1.Envelope {
	return fixture.seedDeliveredAt(t, ctx, sequence, time.Time{})
}

func (fixture *eventGapPGFixture) seedDeliveredAt(
	t *testing.T,
	ctx context.Context,
	sequence int64,
	occurredAt time.Time,
) eventsv1.Envelope {
	t.Helper()
	eventID := uuid.New()
	seedPublisherEvent(t, ctx, fixture.publisherPGFixture, publisherEventSeed{
		eventID: eventID, tenantID: fixture.tenantID,
		aggregateID: fixture.aggregate, sequence: sequence, occurredAt: occurredAt,
	})
	authority := publisherAuthorityFor(t, ctx, fixture.publisherPGFixture, fixture.tenantID)
	claim := claimPipelineEvent(t, ctx, fixture.publisherPGFixture, authority, eventID, 1)
	confirm := eventGapDigest("confirm:" + eventID.String())
	if _, err := fixture.store.ConfirmPublish(ctx, claim, confirm); err != nil {
		t.Fatalf("confirm delivered event %d: %v", sequence, err)
	}
	fixture.envelopes[sequence] = claim.Envelope()
	return claim.Envelope()
}

func (fixture eventGapPGFixture) deliver(
	t *testing.T,
	ctx context.Context,
	sequence int64,
) brokerrabbit.HandleOutcome {
	t.Helper()
	outcome, err := fixture.handler.Handle(
		ctx, fixture.envelopes[sequence], brokerrabbit.DeliveryMetadata{},
	)
	if err != nil {
		t.Fatalf("deliver event %d: %v", sequence, err)
	}
	return outcome
}

func (fixture eventGapPGFixture) gapIDs(
	t *testing.T,
	ctx context.Context,
	sequence int64,
) (uuid.UUID, uuid.UUID) {
	t.Helper()
	var rootID, jobID uuid.UUID
	err := fixture.owner.QueryRowContext(ctx, `
SELECT root.id,job.id
FROM platform.dead_letters AS root
JOIN platform.jobs AS job
  ON job.payload_ref='deadletter:event-gap:'||root.id::text
WHERE root.record_kind='root' AND root.reason_class='sequence_gap'
  AND root.source_id=$1`, fixture.envelopes[sequence].EventID).Scan(&rootID, &jobID)
	if err != nil {
		t.Fatalf("read event %d gap identities: %v", sequence, err)
	}
	return rootID, jobID
}

func (fixture eventGapPGFixture) gapReady(
	t *testing.T,
	ctx context.Context,
	jobID uuid.UUID,
) bool {
	t.Helper()
	var ready bool
	if err := fixture.runtime.QueryRow(
		ctx, "SELECT platform.platform_async_event_gap_job_ready_v2($1)", jobID,
	).Scan(&ready); err != nil {
		t.Fatalf("read gap readiness: %v", err)
	}
	return ready
}

func (fixture eventGapPGFixture) jobAuthority(
	t *testing.T,
	ctx context.Context,
) JobAuthority {
	t.Helper()
	token, err := fixture.store.BeginJobAuthorityPage(256)
	if err != nil {
		t.Fatal(err)
	}
	page, err := fixture.store.ListJobAuthorities(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	for _, authority := range page.Authorities() {
		if authority.scopeID == fixture.tenantID && authority.groupID == "workflow" &&
			authority.domain == "workflow" {
			return authority
		}
	}
	t.Fatal("event gap job authority was not listed")
	return JobAuthority{}
}

func (fixture eventGapPGFixture) claimGap(
	t *testing.T,
	ctx context.Context,
	want uuid.UUID,
) JobClaim {
	t.Helper()
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.jobAuthority(t, ctx), uuid.New(), 1,
		consumerJobCatalogForPipeline(fixture.consumers),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != want {
		t.Fatalf("claim event gap = %#v, want %s: %v", claims, want, err)
	}
	return claims[0]
}

func eventGapDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

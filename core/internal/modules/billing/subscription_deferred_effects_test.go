package billing

import (
	"errors"
	"testing"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func TestDeferredSubscriptionEffectsSealAtomicWorkerOutcome(t *testing.T) {
	tenantID, job, _ := subscriptionBoundaryFixture(t)
	effects := newDeferredSubscriptionEffects(tenantID)
	artifact, err := effects.EnqueueBoundary(nil, job)
	if err != nil || artifact.PayloadReference == "" {
		t.Fatalf("enqueue boundary = %#v, %v", artifact, err)
	}
	if _, err := effects.EnqueueBoundary(nil, job); !errors.Is(
		err, subscriptionapp.ErrSnapshot,
	) {
		t.Fatalf("duplicate successor error = %v", err)
	}
	cancelledAt := job.RunAt.Add(-time.Minute)
	cancelledJob := uuid.MustParse("019c6a7e-7c00-7000-8000-000000000043")
	if err := effects.CancelBoundary(nil, subscriptionapp.JobCancellation{
		ID: cancelledJob, CancelledAt: cancelledAt, Reason: "superseded",
	}); err != nil {
		t.Fatal(err)
	}
	snapshot := subscriptionapp.SubscriptionSnapshot{
		ID: job.SubscriptionID, TenantID: tenantID, ProviderAccountID: uuid.New(),
		OfferingVersionID: uuid.New(), PriceVersionID: uuid.New(), Currency: "TRY",
		State: "active", Revision: 8, UpdatedAt: cancelledAt,
	}
	eventID := uuid.MustParse("019c6a7e-7c00-7000-8000-000000000044")
	causationID := uuid.MustParse("019c6a7e-7c00-7000-8000-000000000045")
	if err := effects.PublishSubscription(nil, subscriptionapp.OutboxIntent{
		ID: eventID, EventType: "billing.subscription.changed", EventVersion: 1,
		SubscriptionID: snapshot.ID, AggregateRevision: snapshot.Revision,
		ActorID: causationID, CausationID: causationID,
		OccurredAt: cancelledAt, ChangeKind: "period_renewed",
	}, snapshot); err != nil {
		t.Fatal(err)
	}
	entitlementEventID := uuid.MustParse("019c6a7e-7c00-7000-8000-000000000046")
	if err := effects.PublishSubscription(nil, subscriptionapp.OutboxIntent{
		ID: entitlementEventID, EventType: "billing.entitlements.changed",
		EventVersion: 1, SubscriptionID: snapshot.ID,
		AggregateRevision: snapshot.Revision, ActorID: causationID,
		CausationID: causationID, OccurredAt: cancelledAt,
		ChangeKind: "period_renewed",
	}, snapshot); err != nil {
		t.Fatal(err)
	}
	outcome, err := effects.outcome()
	if err != nil || !outcome.Valid() {
		t.Fatalf("deferred outcome = %+v, %v", outcome, err)
	}
	successor, ok := outcome.Successor()
	if !ok || successor.JobID != job.ID || successor.RunAt != job.RunAt {
		t.Fatalf("successor = %#v, %v", successor, ok)
	}
	cancellations := outcome.Cancellations()
	if len(cancellations) != 1 || cancellations[0].JobID != cancelledJob {
		t.Fatalf("cancellations = %#v", cancellations)
	}
	events := outcome.Events()
	if len(events) != 2 || events[0].EventID != eventID ||
		events[0].AggregateVersion != snapshot.Revision ||
		events[1].EventID != entitlementEventID ||
		events[1].AggregateType != "billing.entitlements" {
		t.Fatalf("events = %#v", events)
	}
}

func TestDeferredSubscriptionEffectsRejectInvalidCollectedIntent(t *testing.T) {
	effects := newDeferredSubscriptionEffects(uuid.New())
	if err := effects.CancelBoundary(nil, subscriptionapp.JobCancellation{
		ID: uuid.New(), CancelledAt: time.Now(), Reason: "invalid",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := effects.outcome(); !errors.Is(
		err, platformasync.ErrDatabaseJobOutcomeInvalid,
	) {
		t.Fatalf("invalid collected intent error = %v", err)
	}
}

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func TestExecuteWebhookAppliesMappedOrderedSubscriptionFact(t *testing.T) {
	fixture := newProcessorFixture(t, 1, 0, "received")
	fixture.database.rows[findLatestProviderSequenceSQL] = []fixtureRow{{err: errFixtureRowMissing}}
	fixture.database.rows[insertProviderEventSQL] = []fixtureRow{{values: []any{fixture.eventID}}}
	decision, err := fixture.processor.ExecuteWebhook(context.Background(), fixture.receiptID)
	if err != nil || decision.State != providerapp.WebhookApplied {
		t.Fatalf("decision=%+v error=%v", decision, err)
	}
	if fixture.updater.calls != 1 ||
		fixture.updater.request.SubscriptionID != fixture.subscriptionID ||
		fixture.updater.request.ProviderReceiptID != fixture.receiptID ||
		fixture.updater.request.NormalizedFactID != fixture.eventID ||
		len(fixture.jobs.jobs) != 0 {
		t.Fatalf("updater=%+v jobs=%d", fixture.updater.request, len(fixture.jobs.jobs))
	}
}

func TestExecuteWebhookQueuesReconciliationOnOrderingGap(t *testing.T) {
	fixture := newProcessorFixture(t, 2, 1, "received")
	fixture.database.rows[findLatestProviderSequenceSQL] = []fixtureRow{{err: errFixtureRowMissing}}
	fixture.database.rows[insertProviderEventSQL] = []fixtureRow{{values: []any{fixture.eventID}}}
	fixture.database.rows[markWebhookReconcileSQL] = []fixtureRow{{values: []any{fixture.receiptID}}}
	decision, err := fixture.processor.ExecuteWebhook(context.Background(), fixture.receiptID)
	if err != nil || decision.State != providerapp.WebhookReconcileQueued {
		t.Fatalf("decision=%+v error=%v", decision, err)
	}
	if fixture.updater.calls != 0 || len(fixture.jobs.jobs) != 1 ||
		fixture.jobs.jobs[0].Policy() != providerapp.ProviderReconciliationPolicy() ||
		fixture.jobs.ids[0] != fixture.reconciliationID {
		t.Fatalf("updates=%d jobs=%d", fixture.updater.calls, len(fixture.jobs.jobs))
	}
}

func TestExecuteWebhookReplaysOnlyExactTerminalProviderEvent(t *testing.T) {
	fixture := newProcessorFixture(t, 1, 0, "applied")
	digest := fixture.event.FactSHA256()
	fixture.database.rows[findProviderEventSQL] = []fixtureRow{{values: []any{
		fixture.eventID, fixture.event.Schema(), fixture.event.Revision(),
		string(fixture.event.Ordering().Mode()), int64(1), int64(0),
		fixture.event.CanonicalFact(), fixture.event.OccurredAt(), digest[:], "applied",
	}}}
	decision, err := fixture.processor.ExecuteWebhook(context.Background(), fixture.receiptID)
	if err != nil || decision.State != providerapp.WebhookDuplicate {
		t.Fatalf("decision=%+v error=%v", decision, err)
	}
	if fixture.updater.calls != 0 || len(fixture.jobs.jobs) != 0 {
		t.Fatalf("updates=%d jobs=%d", fixture.updater.calls, len(fixture.jobs.jobs))
	}
}

type processorFixture struct {
	processor        *providerapp.Processor
	database         *fixtureDatabase
	updater          *fixtureUpdater
	jobs             *fixtureJobs
	event            providerapp.NormalizedEvent
	tenantID         uuid.UUID
	subscriptionID   uuid.UUID
	receiptID        uuid.UUID
	eventID          uuid.UUID
	reconciliationID uuid.UUID
}

func newProcessorFixture(
	t *testing.T,
	sequence int64,
	previous int64,
	state string,
) processorFixture {
	t.Helper()
	now := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
	tenantID, subscriptionID := uuid.New(), uuid.New()
	receiptID, _ := uuid.NewV7()
	eventID, _ := uuid.NewV7()
	reconciliationID, _ := uuid.NewV7()
	authorityID, mappingID, accountID := uuid.New(), uuid.New(), uuid.New()
	raw := []byte(`{"id":"evt_30","object":"sub_20"}`)
	start, end := now, now.AddDate(0, 1, 0)
	fact := []byte(fmt.Sprintf(
		`{"boundary":null,"grace_ends_at":null,"period_end":%q,"period_start":%q,"state":"active","subscription_id":%q,"trial_ends_at":null}`,
		end.Format(time.RFC3339Nano), start.Format(time.RFC3339Nano), subscriptionID,
	))
	ordering, _ := providerapp.ContiguousOrdering(sequence, previous)
	event, err := providerapp.NewNormalizedEvent(
		subscriptionSnapshotSchema, subscriptionSnapshotRevision,
		now.Add(-time.Second), ordering, fact,
	)
	if err != nil {
		t.Fatal(err)
	}
	registry, _ := fixtureRegistry(t, event)
	database := &fixtureDatabase{rows: map[string][]fixtureRow{
		loadWebhookSQL: {
			{values: fixtureWebhookRows(
				authorityID, mappingID, accountID, subscriptionID, now, raw, state,
			)},
			{values: fixtureWebhookRows(
				authorityID, mappingID, accountID, subscriptionID, now, raw,
				mapLoadState(state),
			)},
		},
		findProviderEventSQL: {{err: errFixtureRowMissing}},
	}}
	if state == "received" {
		database.rows[markWebhookProcessingSQL] = []fixtureRow{{values: []any{receiptID}}}
	}
	updater := &fixtureUpdater{result: subscriptionapp.Result{Subscription: subscriptionapp.SubscriptionSnapshot{
		ID: subscriptionID, TenantID: tenantID,
	}}}
	jobs := &fixtureJobs{}
	identifiers := []uuid.UUID{eventID, reconciliationID}
	storage := &processorStore{
		tenantID: tenantID,
		database: databaseAccess{
			queryRow: database.queryRow,
			execute:  func(context.Context, string, ...any) error { return nil },
			allowed:  func(context.Context) bool { return true },
			notFound: func(err error) bool { return err == errFixtureRowMissing },
		},
		nextID: func() (uuid.UUID, error) {
			identifier := identifiers[0]
			identifiers = identifiers[1:]
			return identifier, nil
		},
		clock: func() time.Time { return now }, jobs: jobs, update: updater,
	}
	processor, err := providerapp.NewProcessor(
		registry, fixtureClock{now: now}, fixtureOpener{body: raw}, storage,
	)
	if err != nil {
		t.Fatal(err)
	}
	return processorFixture{
		processor: processor, database: database, updater: updater, jobs: jobs,
		event: event, tenantID: tenantID, subscriptionID: subscriptionID,
		receiptID: receiptID, eventID: eventID, reconciliationID: reconciliationID,
	}
}

func mapLoadState(state string) string {
	if state == "received" {
		return "processing"
	}
	return state
}

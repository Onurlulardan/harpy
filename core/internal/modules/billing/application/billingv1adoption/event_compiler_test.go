package billingv1adoptionapp

import "testing"

func TestCompileLegacyEventsPreservesExactAuditPayload(t *testing.T) {
	snapshot, mapping := validAdoptionFixture()
	bundle, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	events := bundle.Tenants[0].LegacyEvents
	if len(events) != 1 || events[0].EventType != "payment_succeeded" ||
		events[0].AmountCents != 900 || events[0].Currency != "USD" ||
		string(events[0].Metadata) != `{"receipt":"R-1"}` ||
		events[0].SubscriptionID == nil || events[0].SourceSHA256 == "" {
		t.Fatalf("compiled legacy events = %+v", events)
	}
}

func TestCompileLegacyEventsRejectsUnknownTypeAndOrphanSubscription(t *testing.T) {
	snapshot, mapping := validAdoptionFixture()
	snapshot.Tenants[0].Events[0].EventType = "unknown"
	missing := snapshot.Tenants[0].Events[0].ID
	snapshot.Tenants[0].Events[0].SubscriptionID = &missing
	bundle, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Tenants[0].Eligible || len(bundle.Tenants[0].LegacyEvents) != 0 {
		t.Fatalf("invalid legacy event was accepted: %+v", bundle.Tenants[0])
	}
}

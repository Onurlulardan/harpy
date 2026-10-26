package billing

import (
	"strings"
	"testing"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func TestEntitlementEventUsesIndependentStreamAndRegisteredSchema(t *testing.T) {
	snapshot := subscriptionEventSnapshot(t)
	now := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
	intent := subscriptionapp.OutboxIntent{
		ID: testV7(t), EventType: "billing.entitlements.changed",
		EventVersion: 1, SubscriptionID: snapshot.ID,
		AggregateRevision: snapshot.Revision, ActorID: testV7(t),
		CausationID: testV7(t), OccurredAt: now,
		ChangeKind: "fixture_change",
	}
	payload, digest, schema, aggregate, key, err := subscriptionEventDocument(
		intent, snapshot,
	)
	if err != nil {
		t.Fatal(err)
	}
	if schema != entitlementChangeSchemaSHA256 || aggregate != "billing.entitlements" ||
		len(digest) != 64 || !strings.Contains(string(payload), snapshot.ID.String()) ||
		key != "billing.entitlements.changed:"+snapshot.ID.String()+":4" {
		t.Fatalf("entitlement event document = %s %s %s %s %s", payload, digest, schema, aggregate, key)
	}
}

func subscriptionEventSnapshot(t *testing.T) subscriptionapp.SubscriptionSnapshot {
	t.Helper()
	now := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
	return subscriptionapp.SubscriptionSnapshot{
		ID: uuid.New(), TenantID: uuid.New(), ProviderAccountID: uuid.New(),
		OfferingVersionID: uuid.New(), PriceVersionID: uuid.New(),
		Currency: "TRY", Revision: 4, UpdatedAt: now,
	}
}

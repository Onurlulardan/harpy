package compatibility

import (
	"context"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

type publicDatabaseStub struct {
	resolved compatibilityapp.LegacyMapping
	recorded compatibilityapp.LegacyMapping
	key      compatibilityapp.LegacyKey
	at       time.Time
}

func (stub *publicDatabaseStub) ResolveAnalyticsLegacyPublic(
	_ context.Context, key compatibilityapp.LegacyKey,
) (compatibilityapp.LegacyMapping, error) {
	stub.key = key
	return stub.resolved, nil
}

func (stub *publicDatabaseStub) RecordAnalyticsLegacyPublicHit(
	_ context.Context, route compatibilityapp.LegacyMapping, at time.Time,
) (compatibilityapp.LegacyMapping, error) {
	stub.recorded = route
	stub.at = at
	result := route
	result.HitCount++
	result.LastHitAt = &at
	return result, nil
}

func TestPublicStoreUsesOnlyTypedPlatformOperations(t *testing.T) {
	mapping := legacyMapping()
	database := &publicDatabaseStub{resolved: mapping}
	store, err := NewPublicStore(database)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := store.Resolve(context.Background(), mapping.Key)
	if err != nil || resolved != mapping || database.key != mapping.Key {
		t.Fatalf("resolved=%+v key=%+v error=%v", resolved, database.key, err)
	}
	recorded, err := store.RecordHit(context.Background(), mapping, testTime)
	if err != nil || recorded.HitCount != mapping.HitCount+1 ||
		database.recorded != mapping || !database.at.Equal(testTime) {
		t.Fatalf("recorded=%+v route=%+v at=%s error=%v",
			recorded, database.recorded, database.at, err)
	}
}

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestActorBridgeReconcileReturnsAccountedHitEvidence(t *testing.T) {
	latest := time.Unix(2_000_000_001, 0).UTC()
	queryer := &actorBridgeTestQueryer{row: actorBridgeTestRow{scan: func(values ...any) error {
		*values[0].(*int64) = 3
		setNullTime(values[1], latest, true)
		return nil
	}}}
	database, _ := NewActorBridgeDatabase(queryer)
	result, err := database.ReconcileHits(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	gotLatest, exists := result.LatestHitAt()
	if result.AccountedCount() != 3 || !exists || gotLatest != latest ||
		queryer.query != reconcileIAMV1ActorAssertionHitsSQL || queryer.arguments[0] != 100 {
		t.Fatal("actor bridge reconciliation evidence differs")
	}
}

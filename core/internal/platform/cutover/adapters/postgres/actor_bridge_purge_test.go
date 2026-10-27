package postgres

import (
	"context"
	"testing"
)

func TestActorBridgePurgeIsBounded(t *testing.T) {
	queryer := &actorBridgeTestQueryer{row: actorBridgeTestRow{scan: func(values ...any) error {
		*values[0].(*int64) = 9
		return nil
	}}}
	database, _ := NewActorBridgeDatabase(queryer)
	deleted, err := database.Purge(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 9 || queryer.query != purgeIAMV1ActorAssertionsSQL ||
		queryer.arguments[0] != 10 {
		t.Fatal("actor bridge purge result differs")
	}
	if _, err := database.Purge(context.Background(), 0); err == nil {
		t.Fatal("unbounded actor bridge purge was accepted")
	}
}

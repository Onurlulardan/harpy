//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
)

func assertActorBridgeConcurrentReplay(
	t *testing.T,
	ctx context.Context,
	dsn string,
	authority actorbridge.Authority,
) {
	t.Helper()
	firstConnection := connectActorBridgeV1(t, ctx, dsn, actorBridgeSchemaA)
	defer firstConnection.Close(ctx)
	secondConnection := connectActorBridgeV1(t, ctx, dsn, actorBridgeSchemaA)
	defer secondConnection.Close(ctx)
	firstTransaction, err := firstConnection.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	secondTransaction, err := secondConnection.Begin(ctx)
	if err != nil {
		_ = firstTransaction.Rollback(ctx)
		t.Fatal(err)
	}
	claims := issueActorBridgeClaims(t, time.Now(), actorBridgeTenantA, authority)
	first := newActorBridgeTestConsumer(firstTransaction)
	if _, err := first.Consume(ctx, claims); err != nil {
		t.Fatalf("first concurrent consume: %v", err)
	}
	second := newActorBridgeTestConsumer(secondTransaction)
	started := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		close(started)
		_, consumeErr := second.Consume(ctx, claims)
		result <- consumeErr
	}()
	<-started
	time.Sleep(50 * time.Millisecond)
	if err := firstTransaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-result; !actorbridge.FaultIs(err, actorbridge.FaultReplay) {
		t.Fatalf("second concurrent consume error = %v", err)
	}
	if err := secondTransaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
}

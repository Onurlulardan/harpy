//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func assertConsumerStoreCases(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	producers, consumers := consumerCatalogs(t, eventsv1.EffectDatabaseTransactional)
	repository := consumerRepositoryFixture(producers, consumers)
	connection := consumerRuntimeConnection(t, ctx, database)
	t.Run("rollback has no partial effect", func(t *testing.T) {
		assertConsumerRollback(t, ctx, connection, repository)
	})
	t.Run("reorder parks before effect", func(t *testing.T) {
		assertConsumerGap(t, ctx, connection, repository)
	})
	t.Run("cross tenant fails closed", func(t *testing.T) {
		assertConsumerCrossTenant(t, ctx, connection, repository)
	})
	t.Run("concurrent duplicate runs once", func(t *testing.T) {
		assertConcurrentConsumerDuplicate(t, ctx, database, connection, repository)
	})
	t.Run("bootstrap establishes next sequence", func(t *testing.T) {
		assertConsumerBootstrap(t, ctx, connection, repository)
	})
	t.Run("external receipt must already be durable", func(t *testing.T) {
		assertConsumerExternalReceipt(t, ctx, connection)
	})
	t.Run("stale producer stops before effect", func(t *testing.T) {
		assertConsumerStaleProducer(t, ctx, database, connection)
	})
}

func assertConsumerRollback(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	repository *consumerRepository,
) {
	t.Helper()
	envelope := consumerEnvelope(
		t, repository.producers, "71000000-0000-4000-8000-000000000001",
		"70000000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	seedConsumerEvent(t, ctx, connection, envelope)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status, _ := inspectConsumerEvent(t, ctx, transaction, repository, envelope); status != "apply" {
		t.Fatalf("rollback event status = %s", status)
	}
	insertConsumerProbe(t, ctx, transaction, envelope.EventID, "rollback")
	completeConsumerEvent(
		t, ctx, transaction, repository, envelope,
		"71000000-0000-4000-8000-000000000002",
	)
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertConsumerCounts(t, ctx, connection, envelope.EventID, 0, 0)
	transaction, err = connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status, expected := inspectConsumerEvent(
		t, ctx, transaction, repository, envelope,
	); status != "apply" || expected != 1 {
		t.Fatalf("post-rollback decision = %s/%d", status, expected)
	}
}

func assertConsumerGap(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	repository *consumerRepository,
) {
	t.Helper()
	aggregateID := "72000000-0000-4000-8000-000000000001"
	eventOne := consumerEnvelope(
		t, repository.producers, "72100000-0000-4000-8000-000000000001",
		aggregateID, 1, consumerTenantA,
	)
	eventTwo := consumerEnvelope(
		t, repository.producers, "72200000-0000-4000-8000-000000000001",
		aggregateID, 2, consumerTenantA,
	)
	eventThree := consumerEnvelope(
		t, repository.producers, "72300000-0000-4000-8000-000000000001",
		aggregateID, 3, consumerTenantA,
	)
	for _, envelope := range []eventsv1.Envelope{eventOne, eventThree, eventTwo} {
		seedConsumerEvent(t, ctx, connection, envelope)
	}
	consumeConsumerEvent(t, ctx, connection, repository, eventOne, "gap-one")
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	status, expected := inspectConsumerEvent(t, ctx, transaction, repository, eventThree)
	if status != "gap" || expected != 2 {
		t.Fatalf("reordered event decision = %s/%d", status, expected)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertConsumerCounts(t, ctx, connection, eventThree.EventID, 0, 0)
	consumeConsumerEvent(t, ctx, connection, repository, eventTwo, "gap-two")
	consumeConsumerEvent(t, ctx, connection, repository, eventThree, "gap-three")
}

func assertConsumerCrossTenant(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	repository *consumerRepository,
) {
	t.Helper()
	envelope := consumerEnvelope(
		t, repository.producers, "73000000-0000-4000-8000-000000000001",
		"73100000-0000-4000-8000-000000000001", 1, consumerTenantB,
	)
	seedConsumerEvent(t, ctx, connection, envelope)
	if repository.authority.matchesConsumerScope(envelope.TenantID) {
		t.Fatal("tenant B envelope matched tenant A repository")
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status, _ := inspectConsumerEvent(t, ctx, transaction, repository, envelope); status != "producer_missing" {
		t.Fatalf("cross-tenant SQL status = %s", status)
	}
}

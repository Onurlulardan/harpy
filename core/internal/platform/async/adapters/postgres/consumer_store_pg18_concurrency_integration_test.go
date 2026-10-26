//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

type concurrentConsumerResult struct {
	status   string
	expected int64
	err      error
}

func assertConcurrentConsumerDuplicate(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	firstConnection *sql.Conn,
	repository *consumerRepository,
) {
	t.Helper()
	envelope := consumerEnvelope(
		t, repository.producers, "74000000-0000-4000-8000-000000000001",
		"74100000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	seedConsumerEvent(t, ctx, firstConnection, envelope)
	first, err := firstConnection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Rollback()
	if status, _ := inspectConsumerEvent(t, ctx, first, repository, envelope); status != "apply" {
		t.Fatalf("first concurrent decision = %s", status)
	}
	secondConnection := consumerRuntimeConnection(t, ctx, database)
	started := make(chan struct{})
	result := make(chan concurrentConsumerResult, 1)
	go inspectConcurrentDuplicate(
		ctx, secondConnection, repository, envelope, started, result,
	)
	<-started
	select {
	case early := <-result:
		_ = first.Rollback()
		t.Fatalf("second consumer bypassed stream lock: %+v", early)
	case <-time.After(100 * time.Millisecond):
	}
	insertConsumerProbe(t, ctx, first, envelope.EventID, "concurrent")
	completeConsumerEvent(
		t, ctx, first, repository, envelope,
		"74200000-0000-4000-8000-000000000001",
	)
	if err := first.Commit(); err != nil {
		t.Fatal(err)
	}
	second := <-result
	if second.err != nil || second.status != "duplicate" || second.expected != 2 {
		t.Fatalf("second concurrent decision = %+v", second)
	}
	assertConsumerCounts(t, ctx, firstConnection, envelope.EventID, 1, 1)
}

func inspectConcurrentDuplicate(
	ctx context.Context,
	connection *sql.Conn,
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	started chan<- struct{},
	result chan<- concurrentConsumerResult,
) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		close(started)
		result <- concurrentConsumerResult{err: err}
		return
	}
	close(started)
	var value concurrentConsumerResult
	_, value.err = transaction.ExecContext(
		ctx, lockConsumerStreamSQL,
		consumerStreamArguments(repository, envelope.AggregateType, envelope.AggregateID)...,
	)
	if value.err != nil {
		_ = transaction.Rollback()
		result <- value
		return
	}
	value.err = transaction.QueryRowContext(
		ctx, inspectEventSQL, eventInspectionArguments(repository, envelope)...,
	).Scan(&value.status, &value.expected)
	if value.err != nil {
		_ = transaction.Rollback()
		result <- value
		return
	}
	value.err = transaction.Commit()
	result <- value
}

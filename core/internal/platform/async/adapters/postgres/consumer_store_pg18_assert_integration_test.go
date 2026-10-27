//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func consumeConsumerEvent(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	effectKind string,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status, _ := inspectConsumerEvent(t, ctx, transaction, repository, envelope); status != "apply" {
		_ = transaction.Rollback()
		t.Fatalf("consume event status = %s", status)
	}
	insertConsumerProbe(t, ctx, transaction, envelope.EventID, effectKind)
	completeConsumerEvent(t, ctx, transaction, repository, envelope, uuidForEvent(envelope.EventID))
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit consumer event: %v", err)
	}
}

func assertConsumerCounts(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	eventID string,
	wantEffect int,
	wantReceipt int,
) {
	t.Helper()
	var effectCount, receiptCount int
	err := connection.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.consumer_effect_probe WHERE event_id=$1),
 (SELECT count(*) FROM platform.inbox WHERE event_id=$1)`, eventID).
		Scan(&effectCount, &receiptCount)
	if err != nil || effectCount != wantEffect || receiptCount != wantReceipt {
		t.Fatalf("event counts = %d/%d, want %d/%d, error %v",
			effectCount, receiptCount, wantEffect, wantReceipt, err)
	}
}

func uuidForEvent(eventID string) string {
	bytes := []byte(eventID)
	bytes[0], bytes[1] = '8', '8'
	return string(bytes)
}

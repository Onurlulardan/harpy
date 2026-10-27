//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func assertConsumerBootstrap(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	repository *consumerRepository,
) {
	t.Helper()
	receipt := consumerBootstrapReceipt(t, repository)
	_, digest, err := snapshotBootstrapReceipt(receipt, repository.consumers)
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status := inspectConsumerBootstrap(
		t, ctx, transaction, repository, receipt, digest,
	); status != "apply" {
		t.Fatalf("initial bootstrap decision = %s", status)
	}
	insertConsumerProbe(t, ctx, transaction, receipt.ReceiptID, "bootstrap")
	if _, err := transaction.ExecContext(
		ctx, insertBootstrapWatermarkSQL,
		bootstrapInsertArguments(repository, receipt, digest)...,
	); err != nil {
		t.Fatalf("insert bootstrap watermark: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	transaction, err = connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status := inspectConsumerBootstrap(
		t, ctx, transaction, repository, receipt, digest,
	); status != "duplicate" {
		_ = transaction.Rollback()
		t.Fatalf("replayed bootstrap decision = %s", status)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	envelope := consumerEnvelope(
		t, repository.producers, "75000000-0000-4000-8000-000000000001",
		receipt.AggregateID, receipt.ThroughSequence+1, consumerTenantA,
	)
	seedConsumerEvent(t, ctx, connection, envelope)
	consumeConsumerEvent(t, ctx, connection, repository, envelope, "after-bootstrap")
	assertConsumerCounts(t, ctx, connection, envelope.EventID, 1, 1)
}

func inspectConsumerBootstrap(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	repository *consumerRepository,
	receipt eventsv1.BootstrapReceipt,
	digest string,
) string {
	t.Helper()
	if _, err := transaction.ExecContext(
		ctx, lockConsumerStreamSQL,
		consumerStreamArguments(repository, receipt.AggregateType, receipt.AggregateID)...,
	); err != nil {
		t.Fatalf("lock consumer bootstrap stream: %v", err)
	}
	var status string
	err := transaction.QueryRowContext(
		ctx, inspectBootstrapSQL,
		bootstrapArguments(repository, receipt, digest)...,
	).Scan(&status)
	if err != nil {
		t.Fatalf("inspect bootstrap: %v", err)
	}
	if status == "apply" {
		if _, err := transaction.ExecContext(
			ctx, assertConsumerProducerSQL,
			producerAuthorityArguments(receipt.ProducerAuthority)...,
		); err != nil {
			t.Fatalf("assert bootstrap producer: %v", err)
		}
	}
	return status
}

func consumerBootstrapReceipt(
	t *testing.T,
	repository *consumerRepository,
) eventsv1.BootstrapReceipt {
	t.Helper()
	tenant := consumerTenantA
	value, err := eventsv1.NewBootstrapReceipt(eventsv1.BootstrapReceipt{
		ReceiptID: "75100000-0000-4000-8000-000000000001",
		Consumer:  repository.consumer, ConsumerVersion: repository.consumerVersion,
		TenantID: &tenant, ProducerAuthority: consumerEventAuthority(tenant),
		ConsumerAuthority: repository.authority.eventAuthority(),
		AggregateType:     "workflow.instance",
		AggregateID:       "75200000-0000-4000-8000-000000000001",
		ThroughSequence:   5, StateReference: "projection/workflow/bootstrap",
		StateSHA256: strings.Repeat("e", 64), CorrelationID: "consumer:bootstrap.1",
		CausationID: consumerCause, OccurredAt: "2026-07-15T10:00:00Z",
	}, repository.consumers)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

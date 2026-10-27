//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

type consumerSQLExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func seedConsumerEvent(
	t *testing.T,
	ctx context.Context,
	executor consumerSQLExecutor,
	envelope eventsv1.Envelope,
) {
	t.Helper()
	authority := envelope.ProducerAuthority
	result, err := executor.ExecContext(ctx, enqueueEventSQL,
		envelope.EventID, *envelope.TenantID, string(authority.ScopeKind), authority.ScopeID,
		authority.GroupID, authority.GroupGeneration, authority.Domain,
		authority.CutoverGeneration, string(authority.WriterOrigin), envelope.ContractVersion,
		envelope.EventType, envelope.EventVersion, envelope.AggregateType,
		envelope.AggregateID, envelope.AggregateVersion, envelope.EventSequence,
		envelope.OccurredAt, envelope.ActorID, envelope.CorrelationID,
		envelope.CausationID, envelope.IdempotencyKey,
		envelope.BusinessEffectIdempotencyKey, []byte(envelope.PayloadClone()),
		envelope.PayloadSHA256, envelope.PayloadSchemaSHA256,
	)
	if err != nil {
		t.Fatalf("seed consumer event: %v", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		t.Fatalf("seed consumer event rows = %d", rows)
	}
}

func inspectConsumerEvent(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	repository *consumerRepository,
	envelope eventsv1.Envelope,
) (string, int64) {
	t.Helper()
	if _, err := transaction.ExecContext(
		ctx, lockConsumerStreamSQL,
		consumerStreamArguments(repository, envelope.AggregateType, envelope.AggregateID)...,
	); err != nil {
		t.Fatalf("lock consumer event stream: %v", err)
	}
	var status string
	var expected int64
	err := transaction.QueryRowContext(
		ctx, inspectEventSQL, eventInspectionArguments(repository, envelope)...,
	).Scan(&status, &expected)
	if err != nil {
		t.Fatalf("inspect consumer event: %v", err)
	}
	if status == "apply" {
		_, err = transaction.ExecContext(
			ctx, assertConsumerProducerSQL,
			producerAuthorityArguments(envelope.ProducerAuthority)...,
		)
		if err != nil {
			t.Fatalf("assert consumer event producer: %v", err)
		}
	}
	return status, expected
}

func completeConsumerEvent(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	receiptID string,
) {
	t.Helper()
	receipt := integrationEventReceipt(t, repository, envelope, receiptID)
	receipt, digest, err := snapshotEventReceipt(
		receipt, envelope, repository.producers, repository.consumers,
	)
	if err != nil {
		t.Fatal(err)
	}
	arguments := eventReceiptArguments(repository, envelope, receipt, digest)
	if _, err := transaction.ExecContext(ctx, insertEventReceiptSQL, arguments...); err != nil {
		t.Fatalf("insert integration event receipt: %v", err)
	}
	watermarkArguments := eventWatermarkArguments(
		repository, envelope, digest, uuid.New(),
	)
	if _, err := transaction.ExecContext(
		ctx, insertEventWatermarkSQL, watermarkArguments...,
	); err != nil {
		t.Fatalf("insert integration event watermark: %v", err)
	}
}

func integrationEventReceipt(
	t *testing.T,
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	receiptID string,
) eventsv1.EffectReceipt {
	t.Helper()
	value, err := eventsv1.NewEffectReceipt(eventsv1.EffectReceipt{
		ReceiptID: receiptID, Consumer: repository.consumer,
		ConsumerVersion: repository.consumerVersion,
		EventID:         envelope.EventID, EventType: envelope.EventType,
		EventVersion: envelope.EventVersion, TenantID: envelope.TenantID,
		ProducerAuthority: envelope.ProducerAuthority,
		ConsumerAuthority: repository.authority.eventAuthority(),
		AggregateType:     envelope.AggregateType, AggregateID: envelope.AggregateID,
		ThroughSequence:   envelope.EventSequence,
		BusinessEffectKey: envelope.BusinessEffectIdempotencyKey,
		ResultReference:   "probe/" + envelope.EventID,
		ResultSHA256:      strings.Repeat("d", 64), CorrelationID: envelope.CorrelationID,
		CausationID: envelope.EventID, OccurredAt: "2026-07-15T10:00:02Z",
	}, envelope, repository.producers, repository.consumers)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func insertConsumerProbe(
	t *testing.T,
	ctx context.Context,
	executor consumerSQLExecutor,
	eventID string,
	effectKind string,
) {
	t.Helper()
	_, err := executor.ExecContext(ctx, `
INSERT INTO platform.consumer_effect_probe (event_id,tenant_id,effect_kind)
VALUES ($1,$2,$3)`, eventID, consumerTenantA, effectKind)
	if err != nil {
		t.Fatalf("insert consumer effect probe: %v", err)
	}
}

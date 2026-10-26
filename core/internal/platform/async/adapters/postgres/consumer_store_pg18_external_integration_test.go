//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func assertConsumerExternalReceipt(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	producers, consumers := consumerCatalogs(t, eventsv1.EffectExternal)
	repository := consumerRepositoryFixture(producers, consumers)
	envelope := consumerEnvelope(
		t, producers, "76000000-0000-4000-8000-000000000001",
		"76100000-0000-4000-8000-000000000001", 1, consumerTenantA,
	)
	external := consumerExternalReceipt(t, envelope)
	var durable bool
	err := connection.QueryRowContext(
		ctx, verifyExternalEffectReceiptSQL,
		consumerExternalReceiptArguments(repository, external)...,
	).Scan(&durable)
	if err != nil || durable {
		t.Fatalf("missing external receipt durability = %t, %v", durable, err)
	}
	persistExternalEffectReceipt(t, ctx, connection, external)
	seedConsumerEvent(t, ctx, connection, envelope)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if status, _ := inspectConsumerEvent(t, ctx, transaction, repository, envelope); status != "apply" {
		t.Fatalf("external event decision = %s", status)
	}
	if err := transaction.QueryRowContext(
		ctx, verifyExternalEffectReceiptSQL,
		consumerExternalReceiptArguments(repository, external)...,
	).Scan(&durable); err != nil || !durable {
		t.Fatalf("durable external receipt verification = %t, %v", durable, err)
	}
	eventReceipt := consumerEventReceipt(
		t, envelope, producers, consumers, &external.ReceiptID,
	)
	insertConsumerProbe(t, ctx, transaction, envelope.EventID, "external-finalize")
	insertExternalEventPair(t, ctx, transaction, repository, envelope, eventReceipt)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertConsumerCounts(t, ctx, connection, envelope.EventID, 1, 1)
}

func persistExternalEffectReceipt(
	t *testing.T,
	ctx context.Context,
	executor consumerSQLExecutor,
	value effectsv1.Receipt,
) {
	t.Helper()
	_, err := executor.ExecContext(ctx, `
INSERT INTO platform.idempotency_receipts (
 receipt_id,record_kind,tenant_id,scope_kind,scope_id,
 authority_group_id,authority_group_generation,authority_domain,
 cutover_generation,writer_origin,request_sha256,retention_class,
 effect_id,effect_kind,effect_revision,effect_key,provider,operation,
 correlation_id,causation_id,effect_status,effect_error_class,effect_code,
 provider_reference,effect_result_ref,effect_result_sha256,completed_at,retain_until
) VALUES (
 $1,'effect',$2,'tenant',$2,$3,$4,$5,$6,'v2',$7,'external_effect',
 $8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,
 $22::timestamptz,TIMESTAMPTZ '2026-10-15 10:00:00+00'
)`, value.ReceiptID, *value.TenantID, value.Authority.GroupID,
		value.Authority.GroupGeneration, value.Authority.Domain,
		value.Authority.CutoverGeneration, strings.Repeat("a", 64),
		value.EffectID, value.EffectKind, value.EffectRevision, value.IdempotencyKey,
		value.Provider, value.Operation, value.CorrelationID, value.CausationID,
		string(value.Status), optionalErrorClass(value.ErrorClass), value.Code,
		optionalString(value.ProviderReference), optionalString(value.ResultReference),
		optionalString(value.ResultSHA256), value.CompletedAt)
	if err != nil {
		t.Fatalf("persist external effect receipt: %v", err)
	}
}

func insertExternalEventPair(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	repository *consumerRepository,
	envelope eventsv1.Envelope,
	receipt eventsv1.EffectReceipt,
) {
	t.Helper()
	receipt, digest, err := snapshotEventReceipt(
		receipt, envelope, repository.producers, repository.consumers,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(
		ctx, insertEventReceiptSQL,
		eventReceiptArguments(repository, envelope, receipt, digest)...,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(
		ctx, insertEventWatermarkSQL,
		eventWatermarkArguments(repository, envelope, digest, uuid.New())...,
	); err != nil {
		t.Fatal(err)
	}
}

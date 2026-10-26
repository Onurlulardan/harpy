//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertConsumerStaleProducer(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
) {
	t.Helper()
	producers, consumers := staleProducerCatalogs(t)
	repository := consumerRepositoryFixture(producers, consumers)
	envelope := staleProducerEnvelope(t, producers)
	seedConsumerEvent(t, ctx, connection, envelope)
	_, err := database.ExecContext(ctx, `
UPDATE platform.domain_cutovers
SET generation = 2
WHERE scope_kind = 'tenant' AND scope_id = $1
  AND domain = 'notifications'`, consumerTenantA)
	if err != nil {
		t.Fatalf("advance producer generation: %v", err)
	}
	t.Cleanup(func() {
		_, _ = database.ExecContext(context.Background(), `
UPDATE platform.domain_cutovers SET generation = 1
WHERE scope_kind = 'tenant' AND scope_id = $1
  AND domain = 'notifications'`, consumerTenantA)
	})
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(
		ctx, lockConsumerStreamSQL,
		consumerStreamArguments(repository, envelope.AggregateType, envelope.AggregateID)...,
	); err != nil {
		t.Fatalf("lock consumer stream: %v", err)
	}
	var status string
	var expected int64
	err = transaction.QueryRowContext(
		ctx, inspectEventSQL, eventInspectionArguments(repository, envelope)...,
	).Scan(&status, &expected)
	if err != nil || status != "apply" || expected != 1 {
		t.Fatalf("pre-authority decision = %s/%d: %v", status, expected, err)
	}
	_, err = transaction.ExecContext(
		ctx, assertConsumerProducerSQL,
		producerAuthorityArguments(envelope.ProducerAuthority)...,
	)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "P5602" {
		t.Fatalf("stale producer error = %v", err)
	}
	_ = transaction.Rollback()
	assertConsumerCounts(t, ctx, connection, envelope.EventID, 0, 0)
}

func staleProducerCatalogs(
	t *testing.T,
) (eventsv1.ProducerCatalog, eventsv1.ConsumerCatalog) {
	t.Helper()
	producer, err := eventsv1.NewProducerPolicy(
		"workflow.changed", 1, "notifications", strings.Repeat("b", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	producers, err := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{producer})
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := eventsv1.NewConsumerPolicy(
		"projection.workflow", 1, "workflow", "workflow.changed", 1,
		"projection.workflow.changed.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	consumers, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{consumer})
	if err != nil {
		t.Fatal(err)
	}
	return producers, consumers
}

func staleProducerEnvelope(
	t *testing.T,
	producers eventsv1.ProducerCatalog,
) eventsv1.Envelope {
	t.Helper()
	tenant := consumerTenantA
	payload := []byte(`{"ok":true}`)
	value, err := eventsv1.NewEnvelope(eventsv1.Envelope{
		EventID:   "77000000-0000-4000-8000-000000000001",
		EventType: "workflow.changed", EventVersion: 1, TenantID: &tenant,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "communications", GroupGeneration: 1,
			Domain: "notifications", CutoverGeneration: 1,
			WriterOrigin: eventsv1.WriterV2,
		},
		AggregateType:    "workflow.instance",
		AggregateID:      "77100000-0000-4000-8000-000000000001",
		AggregateVersion: 1, EventSequence: 1,
		OccurredAt: "2026-07-15T10:00:00Z", ActorID: consumerActor,
		CorrelationID: "consumer:stale-producer.1", CausationID: consumerCause,
		IdempotencyKey:               "event/stale-producer",
		BusinessEffectIdempotencyKey: "effect/stale-producer",
		PayloadSchemaSHA256:          strings.Repeat("b", 64),
		PayloadSHA256:                consumerDigest(payload), Payload: payload,
	}, producers)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

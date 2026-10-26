package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repository *consumerRepository) BeginEvent(
	ctx context.Context,
	value eventsv1.Envelope,
) (EventDecision, error) {
	if ctx == nil || !repository.valid() {
		return nil, ErrConsumerInputInvalid
	}
	envelope, err := snapshotEnvelope(value, repository.producers)
	if err != nil || !repository.authority.matchesConsumerScope(envelope.TenantID) {
		return nil, ErrConsumerInputInvalid
	}
	policy, err := repository.consumers.Resolve(
		repository.consumer, repository.consumerVersion,
		envelope.EventType, envelope.EventVersion,
	)
	if err != nil || policy.OwnerDomain != repository.authority.domain {
		return nil, ErrConsumerInputInvalid
	}
	_, err = repository.database.Execute(
		ctx, lockConsumerStreamSQL,
		consumerStreamArguments(
			repository, envelope.AggregateType, envelope.AggregateID,
		)...,
	)
	if err != nil {
		return nil, classifyConsumerDatabaseError("lock event stream", err)
	}
	var status string
	var expected int64
	err = repository.database.QueryRow(
		ctx, inspectEventSQL, eventInspectionArguments(repository, envelope)...,
	).Scan(&status, &expected)
	if err != nil {
		return nil, classifyConsumerDatabaseError("inspect event", err)
	}
	if status == "apply" || status == "gap" {
		_, err = repository.database.Execute(
			ctx, assertConsumerProducerSQL,
			producerAuthorityArguments(envelope.ProducerAuthority)...,
		)
		if err != nil {
			return nil, classifyConsumerDatabaseError("assert event producer", err)
		}
	}
	if status == "gap" {
		record, err := newEventGapRecord(repository, envelope, expected)
		if err != nil {
			return nil, fmt.Errorf("build event sequence gap evidence: %w", err)
		}
		parked, err := repository.parkEventGap(ctx, envelope.EventID, record)
		if err != nil {
			return nil, err
		}
		return parked, nil
	}
	return repository.eventDecision(envelope, policy, status, expected)
}

func (repository *consumerRepository) eventDecision(
	envelope eventsv1.Envelope,
	policy eventsv1.ConsumerPolicy,
	status string,
	expected int64,
) (EventDecision, error) {
	switch status {
	case "duplicate":
		return DuplicateEvent{eventID: envelope.EventID}, nil
	case "apply":
		return EventEffectPermit{
			owner: repository.identity, envelope: envelope, policy: policy,
			used: &atomic.Bool{},
		}, nil
	case "producer_missing":
		return nil, ErrProducerEvidenceMissing
	default:
		return nil, fmt.Errorf("inspect event returned %q: %w", status, ErrConsumerInputInvalid)
	}
}

func consumerStreamArguments(
	repository *consumerRepository,
	aggregateType string,
	aggregateID string,
) []any {
	return append(
		consumerAuthorityArguments(repository.authority),
		repository.consumer, repository.consumerVersion, aggregateType, aggregateID,
	)
}

func classifyConsumerDatabaseError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "P5601", "P5602", "P5603":
			return fmt.Errorf("%s: %w", operation, ErrConsumerAuthorityStale)
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func eventInspectionArguments(
	repository *consumerRepository,
	envelope eventsv1.Envelope,
) []any {
	authority := envelope.ProducerAuthority
	return []any{
		"tenant", repository.authority.scopeID,
		repository.consumer, repository.consumerVersion,
		envelope.EventID, *envelope.TenantID,
		string(authority.ScopeKind), authority.ScopeID, authority.GroupID,
		authority.GroupGeneration, authority.Domain, authority.CutoverGeneration,
		string(authority.WriterOrigin), envelope.ContractVersion,
		envelope.EventType, envelope.EventVersion, envelope.AggregateType,
		envelope.AggregateID, envelope.AggregateVersion, envelope.EventSequence,
		envelope.OccurredAt, envelope.ActorID, envelope.CorrelationID,
		envelope.CausationID, envelope.IdempotencyKey,
		envelope.BusinessEffectIdempotencyKey, envelope.PayloadSHA256,
		envelope.PayloadSchemaSHA256,
	}
}

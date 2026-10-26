package postgres

import (
	"context"
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

type eventIdentifiers struct {
	eventID, aggregateID, actorID, causationID uuid.UUID
	occurredAt                                 time.Time
}

func (repository *repository) EnqueueEvent(
	ctx context.Context,
	value eventsv1.Envelope,
) (EnqueueResult, error) {
	if repository == nil || !repository.authority.matchesEvent(value) {
		return EnqueueResult{}, ErrAuthorityMismatch
	}
	if err := value.Validate(repository.producers); err != nil {
		return EnqueueResult{}, fmt.Errorf("validate event envelope: %w", err)
	}
	identifiers, err := parseEventIdentifiers(value)
	if err != nil {
		return EnqueueResult{}, err
	}
	result, err := repository.database.Execute(
		ctx, enqueueEventSQL, identifiers.eventID, repository.authority.scopeID,
		"tenant", repository.authority.scopeID, repository.authority.groupID,
		repository.authority.groupGeneration, repository.authority.domain,
		repository.authority.generation, "v2", value.ContractVersion, value.EventType,
		value.EventVersion, value.AggregateType, identifiers.aggregateID,
		value.AggregateVersion, value.EventSequence, identifiers.occurredAt,
		identifiers.actorID, value.CorrelationID, identifiers.causationID,
		value.IdempotencyKey, value.BusinessEffectIdempotencyKey, []byte(value.PayloadClone()),
		value.PayloadSHA256, value.PayloadSchemaSHA256,
	)
	if err != nil {
		return EnqueueResult{}, fmt.Errorf("insert async event: %w", err)
	}
	if result.RowsAffected() == 1 {
		return EnqueueResult{ID: identifiers.eventID.String(), Created: true}, nil
	}
	return repository.findEventEnqueue(ctx, value, identifiers)
}

func parseEventIdentifiers(value eventsv1.Envelope) (eventIdentifiers, error) {
	values := []string{value.EventID, value.AggregateID, value.ActorID, value.CausationID}
	parsed := make([]uuid.UUID, len(values))
	for index, raw := range values {
		identifier, err := uuid.Parse(raw)
		if err != nil {
			return eventIdentifiers{}, fmt.Errorf("parse event UUID: %w", err)
		}
		parsed[index] = identifier
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, value.OccurredAt)
	if err != nil {
		return eventIdentifiers{}, fmt.Errorf("parse event occurrence time: %w", err)
	}
	return eventIdentifiers{
		eventID: parsed[0], aggregateID: parsed[1], actorID: parsed[2],
		causationID: parsed[3], occurredAt: occurredAt,
	}, nil
}

func (repository *repository) findEventEnqueue(
	ctx context.Context,
	value eventsv1.Envelope,
	identifiers eventIdentifiers,
) (EnqueueResult, error) {
	var existingID string
	var exact bool
	err := repository.database.QueryRow(
		ctx, findEventEnqueueSQL, identifiers.eventID, repository.authority.scopeID,
		value.AggregateType, identifiers.aggregateID, value.EventSequence,
		repository.authority.groupID, repository.authority.groupGeneration,
		repository.authority.domain, repository.authority.generation, value.ContractVersion,
		value.EventType, value.EventVersion, value.AggregateVersion, identifiers.actorID,
		value.CorrelationID, identifiers.causationID, value.IdempotencyKey,
		value.BusinessEffectIdempotencyKey, value.PayloadSHA256, value.PayloadSchemaSHA256,
		identifiers.occurredAt,
	).Scan(&existingID, &exact)
	if err != nil {
		return EnqueueResult{}, fmt.Errorf("read existing async event: %w", err)
	}
	if !exact {
		return EnqueueResult{}, ErrIdempotencyConflict
	}
	return EnqueueResult{ID: existingID}, nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *deadLetterRepository) IngestEvent(
	ctx context.Context,
	root deadletter.EventRoot,
) (deadletter.IngestResult, error) {
	if ctx == nil || !repository.validIngest() {
		return deadletter.IngestResult{}, ErrDeadLetterStoreUnavailable
	}
	consumer := root.Consumer()
	registered, err := repository.consumers.ResolveQueue(consumer.QueueName)
	if err != nil || registered != consumer {
		return deadletter.IngestResult{}, ErrDeadLetterInputInvalid
	}
	envelope := root.Envelope()
	document, err := envelope.MarshalJSON()
	if err != nil || root.RootID() == uuid.Nil || root.ScopeID() == uuid.Nil {
		return deadletter.IngestResult{}, ErrDeadLetterInputInvalid
	}
	decoded, err := deadLetterEventIdentifiers(root, document)
	if err != nil {
		return deadletter.IngestResult{}, err
	}
	return withDeadLetterTransaction(
		ctx, repository,
		func(transaction pgx.Tx) (deadletter.IngestResult, error) {
			authority, err := resolveDeadLetterAuthority(
				ctx, transaction, decoded.scopeKind, root.ScopeID(), consumer.OwnerDomain,
			)
			if err != nil {
				return deadletter.IngestResult{}, err
			}
			return insertDeadLetterEvent(ctx, transaction, root, decoded, authority)
		},
	)
}

type deadLetterEventIdentity struct {
	eventID   uuid.UUID
	tenantID  *uuid.UUID
	scopeKind string
}

func deadLetterEventIdentifiers(
	root deadletter.EventRoot,
	document []byte,
) (deadLetterEventIdentity, error) {
	envelope := root.Envelope()
	eventID, err := uuid.Parse(envelope.EventID)
	if err != nil || eventID == uuid.Nil || len(document) == 0 {
		return deadLetterEventIdentity{}, ErrDeadLetterInputInvalid
	}
	tenantID := root.TenantID()
	scopeKind := string(envelope.ProducerAuthority.ScopeKind)
	if (scopeKind == "tenant" && (tenantID == nil || *tenantID != root.ScopeID())) ||
		(scopeKind == "global" && tenantID != nil) {
		return deadLetterEventIdentity{}, ErrDeadLetterInputInvalid
	}
	if scopeKind != "tenant" && scopeKind != "global" {
		return deadLetterEventIdentity{}, ErrDeadLetterInputInvalid
	}
	return deadLetterEventIdentity{eventID: eventID, tenantID: tenantID, scopeKind: scopeKind}, nil
}

func insertDeadLetterEvent(
	ctx context.Context,
	transaction pgx.Tx,
	root deadletter.EventRoot,
	identity deadLetterEventIdentity,
	authority deadLetterAuthority,
) (deadletter.IngestResult, error) {
	envelope := root.Envelope()
	consumer := root.Consumer()
	var result deadletter.IngestResult
	err := transaction.QueryRow(
		ctx, insertDeadLetterEventSQL,
		root.RootID(), identity.tenantID, identity.scopeKind, root.ScopeID(),
		authority.groupID, authority.groupGeneration, authority.domain, authority.generation,
		identity.eventID, consumer.Consumer, consumer.ConsumerVersion,
		envelope.EventType, envelope.EventVersion, root.SourceSHA256(),
		string(root.ReasonClass()), root.ReasonCode(), root.ErrorSHA256(),
		root.Payload(), envelope.PayloadSHA256, root.RedactedEnvelope(),
	).Scan(&result.RootID, &result.RecordedAt, &result.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		return deadletter.IngestResult{}, ErrDeadLetterInputInvalid
	}
	if err != nil {
		return deadletter.IngestResult{}, fmt.Errorf("persist broker dead-letter: %w", err)
	}
	return result, nil
}

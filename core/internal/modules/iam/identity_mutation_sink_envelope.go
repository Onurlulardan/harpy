package iam

import (
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const identityChangePayloadSchemaSHA256 = "e3c637083d0a126cf82ecd547e147bffa4bda5c9ae19ea7b89135710b1b71e6b"

func (sink *identityMutationSink) identityChangeEnvelope(
	record changeapp.RecordedChange,
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
) (eventsv1.Envelope, error) {
	payload, digest, err := identityChangeEventPayload(record)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	_, producers, err := sink.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	tenantID := scope.TenantID().String()
	operationID := uuid.UUID(record.OperationID).String()
	value := eventsv1.Envelope{
		EventID: operationID, EventType: "iam.identity.changed", EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: identityChangeAuthority(authority),
		AggregateType: "iam.identity." + record.ResourceKind,
		AggregateID:   uuid.UUID(record.AggregateID).String(), AggregateVersion: record.Revision,
		EventSequence: record.EventSequence,
		OccurredAt:    sink.now().UTC().Format(time.RFC3339Nano),
		ActorID:       uuid.UUID(record.ActorID).String(), CorrelationID: operationID,
		CausationID: operationID, IdempotencyKey: operationID,
		BusinessEffectIdempotencyKey: operationID,
		PayloadSchemaSHA256:          identityChangePayloadSchemaSHA256,
		PayloadSHA256:                digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, producers)
}

func identityChangeAuthority(authority platformtx.OwnerAuthority) eventsv1.Authority {
	return eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: authority.ScopeID().String(),
		Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
		WriterOrigin: eventsv1.WriterV2, GroupID: authority.GroupID(),
		GroupGeneration: authority.GroupGeneration(),
	}
}

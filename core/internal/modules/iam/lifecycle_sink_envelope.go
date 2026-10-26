package iam

import (
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (sink *LifecycleChangeSink) envelope(change lifecycle.Change,
	tenantID string, sequence int64,
	authority eventsv1.Authority) (eventsv1.Envelope, error) {
	payload, digest, err := lifecycleEventPayload(change)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	_, producers, err := sink.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	operationID := change.OperationID.String()
	value := eventsv1.Envelope{
		EventID: operationID, EventType: "iam.lifecycle.changed", EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: authority,
		AggregateType: "iam." + change.AggregateKind,
		AggregateID:   change.AggregateID.String(), AggregateVersion: change.Revision,
		EventSequence: sequence, OccurredAt: sink.now().UTC().Format(time.RFC3339Nano),
		ActorID: change.ActorID.String(), CorrelationID: operationID,
		CausationID: operationID, IdempotencyKey: operationID,
		BusinessEffectIdempotencyKey: operationID,
		PayloadSchemaSHA256:          lifecyclePayloadSchemaSHA256,
		PayloadSHA256:                digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, producers)
}

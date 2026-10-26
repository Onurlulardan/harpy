package deadletter

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

type eventMetadata struct {
	ContractVersion              string             `json:"contract_version"`
	EventID                      string             `json:"event_id"`
	EventType                    string             `json:"event_type"`
	EventVersion                 int                `json:"event_version"`
	TenantID                     *string            `json:"tenant_id"`
	ProducerAuthority            eventsv1.Authority `json:"producer_authority"`
	AggregateType                string             `json:"aggregate_type"`
	AggregateID                  string             `json:"aggregate_id"`
	AggregateVersion             int64              `json:"aggregate_version"`
	EventSequence                int64              `json:"event_sequence"`
	OccurredAt                   string             `json:"occurred_at"`
	ActorID                      string             `json:"actor_id"`
	CorrelationID                string             `json:"correlation_id"`
	CausationID                  string             `json:"causation_id"`
	IdempotencyKey               string             `json:"idempotency_key"`
	BusinessEffectIdempotencyKey string             `json:"business_effect_idempotency_key"`
	PayloadSchemaSHA256          string             `json:"payload_schema_sha256"`
	PayloadSHA256                string             `json:"payload_sha256"`
}

func marshalEventMetadata(value eventsv1.Envelope) ([]byte, error) {
	document, err := json.Marshal(eventMetadata{
		ContractVersion: value.ContractVersion, EventID: value.EventID,
		EventType: value.EventType, EventVersion: value.EventVersion, TenantID: value.TenantID,
		ProducerAuthority: value.ProducerAuthority, AggregateType: value.AggregateType,
		AggregateID: value.AggregateID, AggregateVersion: value.AggregateVersion,
		EventSequence: value.EventSequence, OccurredAt: value.OccurredAt,
		ActorID: value.ActorID, CorrelationID: value.CorrelationID,
		CausationID: value.CausationID, IdempotencyKey: value.IdempotencyKey,
		BusinessEffectIdempotencyKey: value.BusinessEffectIdempotencyKey,
		PayloadSchemaSHA256:          value.PayloadSchemaSHA256, PayloadSHA256: value.PayloadSHA256,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal broker dead-letter metadata: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 16_384)
	if err != nil {
		return nil, fmt.Errorf("canonicalize broker dead-letter metadata: %w", err)
	}
	return canonical, nil
}

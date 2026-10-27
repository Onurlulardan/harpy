package eventsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type Envelope struct {
	ContractVersion              string          `json:"contract_version"`
	EventID                      string          `json:"event_id"`
	EventType                    string          `json:"event_type"`
	EventVersion                 int             `json:"event_version"`
	TenantID                     *string         `json:"tenant_id"`
	ProducerAuthority            Authority       `json:"producer_authority"`
	AggregateType                string          `json:"aggregate_type"`
	AggregateID                  string          `json:"aggregate_id"`
	AggregateVersion             int64           `json:"aggregate_version"`
	EventSequence                int64           `json:"event_sequence"`
	OccurredAt                   string          `json:"occurred_at"`
	ActorID                      string          `json:"actor_id"`
	CorrelationID                string          `json:"correlation_id"`
	CausationID                  string          `json:"causation_id"`
	IdempotencyKey               string          `json:"idempotency_key"`
	BusinessEffectIdempotencyKey string          `json:"business_effect_idempotency_key"`
	PayloadSchemaSHA256          string          `json:"payload_schema_sha256"`
	PayloadSHA256                string          `json:"payload_sha256"`
	Payload                      json.RawMessage `json:"payload"`
	policy                       ProducerPolicy
}

func NewEnvelope(value Envelope, catalog ProducerCatalog) (Envelope, error) {
	value.ContractVersion = EnvelopeContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	deterministic, err := canonicaljsonv1.CanonicalizeObject(value.Payload, MaximumCanonicalPayloadBytes)
	if err != nil {
		return Envelope{}, err
	}
	value.Payload = deterministic
	if err := value.Validate(catalog); err != nil {
		return Envelope{}, err
	}
	value.policy, _ = catalog.Resolve(value.EventType, value.EventVersion)
	return value, nil
}

func (value Envelope) Validate(catalog ProducerCatalog) error {
	if value.ContractVersion != EnvelopeContractVersion || !validUUID(value.EventID) ||
		!validType(value.EventType) || value.EventVersion < 1 {
		return fmt.Errorf("event identity is invalid")
	}
	if err := value.ProducerAuthority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	policy, err := catalog.Resolve(value.EventType, value.EventVersion)
	if err != nil || policy.OwnerDomain != value.ProducerAuthority.Domain ||
		policy.PayloadSchemaSHA256 != value.PayloadSchemaSHA256 {
		return fmt.Errorf("event does not match registered producer policy")
	}
	if !validType(value.AggregateType) || !validUUID(value.AggregateID) ||
		value.AggregateVersion < 1 || value.EventSequence < 1 {
		return fmt.Errorf("event aggregate identity is invalid")
	}
	if _, ok := validTimestamp(value.OccurredAt); !ok || !validUUID(value.ActorID) ||
		!validCorrelationID(value.CorrelationID) || !validUUID(value.CausationID) {
		return fmt.Errorf("event audit identity is invalid")
	}
	if !validText(value.IdempotencyKey, MaximumKeyBytes) ||
		!validText(value.BusinessEffectIdempotencyKey, MaximumKeyBytes) {
		return fmt.Errorf("event idempotency key is invalid")
	}
	deterministic, err := canonicaljsonv1.CanonicalizeObject(value.Payload, MaximumCanonicalPayloadBytes)
	if err != nil || !lowerSHA256.MatchString(value.PayloadSHA256) {
		return fmt.Errorf("event payload is invalid")
	}
	digest := sha256.Sum256(deterministic)
	if hex.EncodeToString(digest[:]) != value.PayloadSHA256 {
		return fmt.Errorf("event payload digest does not match deterministic bytes")
	}
	return nil
}

func (value Envelope) PayloadClone() json.RawMessage {
	return append(json.RawMessage(nil), value.Payload...)
}

func (value Envelope) PartitionKey() string {
	return value.ProducerAuthority.ScopeID + "/" + value.AggregateType + "/" + value.AggregateID
}

func (value Envelope) MarshalJSON() ([]byte, error) {
	deterministic, err := canonicaljsonv1.CanonicalizeObject(value.Payload, MaximumCanonicalPayloadBytes)
	if err != nil {
		return nil, err
	}
	value.Payload = deterministic
	if err := value.Validate(ProducerCatalog{policies: []ProducerPolicy{value.policy}}); err != nil {
		return nil, err
	}
	type wire Envelope
	return json.Marshal(wire(value))
}

func DecodeEnvelope(document []byte, catalog ProducerCatalog) (Envelope, error) {
	var value Envelope
	if err := decodeStrict(document, &value); err != nil {
		return Envelope{}, err
	}
	if err := value.Validate(catalog); err != nil {
		return Envelope{}, err
	}
	return NewEnvelope(value, catalog)
}

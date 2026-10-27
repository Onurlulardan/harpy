package eventsv1

import (
	"encoding/json"
	"fmt"
)

type EffectReceipt struct {
	ContractVersion         string    `json:"contract_version"`
	ReceiptID               string    `json:"receipt_id"`
	Consumer                string    `json:"consumer"`
	ConsumerVersion         int       `json:"consumer_version"`
	EventID                 string    `json:"event_id"`
	EventType               string    `json:"event_type"`
	EventVersion            int       `json:"event_version"`
	TenantID                *string   `json:"tenant_id"`
	ProducerAuthority       Authority `json:"producer_authority"`
	ConsumerAuthority       Authority `json:"consumer_authority"`
	AggregateType           string    `json:"aggregate_type"`
	AggregateID             string    `json:"aggregate_id"`
	ThroughSequence         int64     `json:"through_sequence"`
	BusinessEffectKey       string    `json:"business_effect_idempotency_key"`
	ExternalEffectReceiptID *string   `json:"external_effect_receipt_id"`
	ResultReference         string    `json:"result_reference"`
	ResultSHA256            string    `json:"result_sha256"`
	CorrelationID           string    `json:"correlation_id"`
	CausationID             string    `json:"causation_id"`
	OccurredAt              string    `json:"occurred_at"`
	producerPolicy          ProducerPolicy
	consumerPolicy          ConsumerPolicy
	source                  Envelope
}

func NewEffectReceipt(
	value EffectReceipt,
	envelope Envelope,
	producers ProducerCatalog,
	consumers ConsumerCatalog,
) (EffectReceipt, error) {
	value.ContractVersion = EffectReceiptContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	value.ExternalEffectReceiptID = cloneOptional(value.ExternalEffectReceiptID)
	if err := value.Validate(envelope, producers, consumers); err != nil {
		return EffectReceipt{}, err
	}
	value.producerPolicy, _ = producers.Resolve(value.EventType, value.EventVersion)
	value.consumerPolicy, _ = consumers.Resolve(
		value.Consumer, value.ConsumerVersion, value.EventType, value.EventVersion,
	)
	value.source = envelope
	value.source.Payload = envelope.PayloadClone()
	return value, nil
}

func (value EffectReceipt) Validate(
	envelope Envelope,
	producers ProducerCatalog,
	consumers ConsumerCatalog,
) error {
	if err := envelope.Validate(producers); err != nil {
		return fmt.Errorf("event effect source envelope: %w", err)
	}
	if value.ContractVersion != EffectReceiptContractVersion || !validUUID(value.ReceiptID) ||
		!validType(value.Consumer) || value.ConsumerVersion < 1 || !validUUID(value.EventID) ||
		!validType(value.EventType) || value.EventVersion < 1 {
		return fmt.Errorf("event effect receipt identity is invalid")
	}
	if err := value.ProducerAuthority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if err := value.ConsumerAuthority.ValidateConsumer(value.TenantID); err != nil {
		return err
	}
	policy, err := consumers.Resolve(value.Consumer, value.ConsumerVersion, value.EventType, value.EventVersion)
	if err != nil || policy.OwnerDomain != value.ConsumerAuthority.Domain {
		return fmt.Errorf("event effect is not allowed by consumer policy")
	}
	if !value.matchesEnvelope(envelope) {
		return fmt.Errorf("event effect receipt does not match envelope")
	}
	if policy.EffectClass == EffectExternal && !validOptionalUUID(value.ExternalEffectReceiptID) ||
		policy.EffectClass == EffectDatabaseTransactional && value.ExternalEffectReceiptID != nil {
		return fmt.Errorf("event effect receipt class fields are invalid")
	}
	if !validText(value.ResultReference, MaximumReferenceBytes) ||
		!lowerSHA256.MatchString(value.ResultSHA256) || !validCorrelationID(value.CorrelationID) ||
		!validUUID(value.CausationID) {
		return fmt.Errorf("event effect result or audit identity is invalid")
	}
	receiptTime, ok := validTimestamp(value.OccurredAt)
	eventTime, _ := validTimestamp(envelope.OccurredAt)
	if !ok || receiptTime.Before(eventTime) {
		return fmt.Errorf("event effect receipt time is invalid")
	}
	return nil
}

func (value EffectReceipt) matchesEnvelope(envelope Envelope) bool {
	return value.EventID == envelope.EventID && value.EventType == envelope.EventType &&
		value.EventVersion == envelope.EventVersion && optionalEqual(value.TenantID, envelope.TenantID) &&
		value.ProducerAuthority == envelope.ProducerAuthority &&
		value.AggregateType == envelope.AggregateType && value.AggregateID == envelope.AggregateID &&
		value.ThroughSequence == envelope.EventSequence &&
		value.BusinessEffectKey == envelope.BusinessEffectIdempotencyKey &&
		value.CorrelationID == envelope.CorrelationID && value.CausationID == envelope.EventID
}

func (value EffectReceipt) MarshalJSON() ([]byte, error) {
	producers := ProducerCatalog{policies: []ProducerPolicy{value.producerPolicy}}
	consumers := ConsumerCatalog{policies: []ConsumerPolicy{value.consumerPolicy}}
	if err := value.Validate(value.source, producers, consumers); err != nil {
		return nil, err
	}
	type wire EffectReceipt
	return json.Marshal(wire(value))
}

func DecodeEffectReceipt(
	document []byte,
	envelope Envelope,
	producers ProducerCatalog,
	consumers ConsumerCatalog,
) (EffectReceipt, error) {
	var value EffectReceipt
	if err := decodeStrict(document, &value); err != nil {
		return EffectReceipt{}, err
	}
	if err := value.Validate(envelope, producers, consumers); err != nil {
		return EffectReceipt{}, err
	}
	return NewEffectReceipt(value, envelope, producers, consumers)
}

func optionalEqual(left *string, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func validOptionalUUID(value *string) bool { return value != nil && validUUID(*value) }

package eventsv1

import (
	"encoding/json"
	"fmt"
)

// SequenceGapEvidence is the exact durable input for ordered-event recovery.
// It contains the original canonical envelope; no payload reconstruction is
// permitted during replay.
type SequenceGapEvidence struct {
	ContractVersion  string   `json:"contract_version"`
	Consumer         string   `json:"consumer"`
	ConsumerVersion  int      `json:"consumer_version"`
	ExpectedSequence int64    `json:"expected_sequence"`
	ReceivedSequence int64    `json:"received_sequence"`
	Envelope         Envelope `json:"envelope"`
	producer         ProducerPolicy
	consumer         ConsumerPolicy
}

func NewSequenceGapEvidence(
	envelope Envelope,
	consumer string,
	consumerVersion int,
	expected int64,
	producers ProducerCatalog,
	consumers ConsumerCatalog,
) (SequenceGapEvidence, error) {
	if err := envelope.Validate(producers); err != nil {
		return SequenceGapEvidence{}, err
	}
	producerPolicy, _ := producers.Resolve(envelope.EventType, envelope.EventVersion)
	snapshot := envelope
	snapshot.Payload = envelope.PayloadClone()
	if envelope.TenantID != nil {
		tenantID := *envelope.TenantID
		snapshot.TenantID = &tenantID
	}
	snapshot.policy = producerPolicy
	consumerPolicy, err := consumers.Resolve(
		consumer, consumerVersion, snapshot.EventType, snapshot.EventVersion,
	)
	if err != nil {
		return SequenceGapEvidence{}, fmt.Errorf("resolve gap consumer: %w", err)
	}
	value := SequenceGapEvidence{
		ContractVersion: SequenceGapContractVersion,
		Consumer:        consumer, ConsumerVersion: consumerVersion,
		ExpectedSequence: expected, ReceivedSequence: snapshot.EventSequence,
		Envelope: snapshot, producer: producerPolicy, consumer: consumerPolicy,
	}
	if err := value.validate(producers, consumers); err != nil {
		return SequenceGapEvidence{}, err
	}
	return value, nil
}

func (value SequenceGapEvidence) validate(
	producers ProducerCatalog,
	consumers ConsumerCatalog,
) error {
	if value.ContractVersion != SequenceGapContractVersion ||
		value.ExpectedSequence < 1 || value.ReceivedSequence != value.Envelope.EventSequence ||
		value.ExpectedSequence >= value.ReceivedSequence {
		return fmt.Errorf("event sequence gap is invalid")
	}
	if err := value.Envelope.Validate(producers); err != nil {
		return err
	}
	policy, err := consumers.Resolve(
		value.Consumer, value.ConsumerVersion,
		value.Envelope.EventType, value.Envelope.EventVersion,
	)
	if err != nil || policy.OwnerDomain == "" {
		return fmt.Errorf("event sequence gap consumer is invalid")
	}
	return nil
}

func (value SequenceGapEvidence) MarshalJSON() ([]byte, error) {
	producers, _ := NewProducerCatalog([]ProducerPolicy{value.producer})
	consumers, _ := NewConsumerCatalog([]ConsumerPolicy{value.consumer})
	if err := value.validate(producers, consumers); err != nil {
		return nil, err
	}
	type wire SequenceGapEvidence
	return json.Marshal(wire(value))
}

func DecodeSequenceGapEvidence(
	document []byte,
	producers ProducerCatalog,
	consumers ConsumerCatalog,
) (SequenceGapEvidence, error) {
	var value SequenceGapEvidence
	if err := decodeStrict(document, &value); err != nil {
		return SequenceGapEvidence{}, err
	}
	return NewSequenceGapEvidence(
		value.Envelope, value.Consumer, value.ConsumerVersion,
		value.ExpectedSequence, producers, consumers,
	)
}

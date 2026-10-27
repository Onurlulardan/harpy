package postgres

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

type eventGapMetadata struct {
	ContractVersion  string             `json:"contract_version"`
	Consumer         string             `json:"consumer"`
	ConsumerVersion  int                `json:"consumer_version"`
	ExpectedSequence int64              `json:"expected_sequence"`
	ReceivedSequence int64              `json:"received_sequence"`
	EventID          string             `json:"event_id"`
	EventType        string             `json:"event_type"`
	EventVersion     int                `json:"event_version"`
	Producer         eventsv1.Authority `json:"producer_authority"`
	AggregateType    string             `json:"aggregate_type"`
	AggregateID      string             `json:"aggregate_id"`
	ActorID          string             `json:"actor_id"`
	PayloadSHA256    string             `json:"payload_sha256"`
}

func marshalEventGapMetadata(
	value eventsv1.SequenceGapEvidence,
) ([]byte, error) {
	envelope := value.Envelope
	document, err := json.Marshal(eventGapMetadata{
		ContractVersion: value.ContractVersion,
		Consumer:        value.Consumer, ConsumerVersion: value.ConsumerVersion,
		ExpectedSequence: value.ExpectedSequence,
		ReceivedSequence: value.ReceivedSequence,
		EventID:          envelope.EventID, EventType: envelope.EventType,
		EventVersion: envelope.EventVersion, Producer: envelope.ProducerAuthority,
		AggregateType: envelope.AggregateType, AggregateID: envelope.AggregateID,
		ActorID: envelope.ActorID, PayloadSHA256: envelope.PayloadSHA256,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal event gap metadata: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 16_384)
	if err != nil {
		return nil, fmt.Errorf("canonicalize event gap metadata: %w", err)
	}
	return canonical, nil
}

package eventsv1

import (
	"encoding/json"
	"fmt"
)

type ProducerPolicy struct {
	ContractVersion         string `json:"contract_version"`
	EventType               string `json:"event_type"`
	EventVersion            int    `json:"event_version"`
	OwnerDomain             string `json:"owner_domain"`
	PayloadSchemaSHA256     string `json:"payload_schema_sha256"`
	MaximumPayloadBytes     int    `json:"maximum_payload_bytes"`
	DeliveredRetentionDays  int    `json:"delivered_retention_days"`
	UndeliveredAlarmSeconds int    `json:"undelivered_alarm_seconds"`
}

func NewProducerPolicy(
	eventType string,
	eventVersion int,
	ownerDomain string,
	payloadSchemaSHA256 string,
) (ProducerPolicy, error) {
	value := ProducerPolicy{
		ContractVersion: ProducerPolicyContractVersion, EventType: eventType,
		EventVersion: eventVersion, OwnerDomain: ownerDomain, PayloadSchemaSHA256: payloadSchemaSHA256,
		MaximumPayloadBytes:     MaximumCanonicalPayloadBytes,
		DeliveredRetentionDays:  DeliveredOutboxRetentionDays,
		UndeliveredAlarmSeconds: UndeliveredOutboxAlarmSeconds,
	}
	if err := value.Validate(); err != nil {
		return ProducerPolicy{}, err
	}
	return value, nil
}

func (value ProducerPolicy) Validate() error {
	if value.ContractVersion != ProducerPolicyContractVersion || !validType(value.EventType) ||
		value.EventVersion < 1 || !validDomain(value.OwnerDomain) ||
		!lowerSHA256.MatchString(value.PayloadSchemaSHA256) ||
		value.MaximumPayloadBytes != MaximumCanonicalPayloadBytes ||
		value.DeliveredRetentionDays != DeliveredOutboxRetentionDays ||
		value.UndeliveredAlarmSeconds != UndeliveredOutboxAlarmSeconds {
		return fmt.Errorf("event producer policy is invalid")
	}
	return nil
}

func (value ProducerPolicy) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire ProducerPolicy
	return json.Marshal(wire(value))
}

func DecodeProducerPolicy(document []byte) (ProducerPolicy, error) {
	var value ProducerPolicy
	if err := decodeStrict(document, &value); err != nil {
		return ProducerPolicy{}, err
	}
	return value, value.Validate()
}

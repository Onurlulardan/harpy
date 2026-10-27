package eventsv1

import (
	"encoding/json"
	"fmt"
)

type ConsumerPolicy struct {
	ContractVersion             string      `json:"contract_version"`
	Consumer                    string      `json:"consumer"`
	ConsumerVersion             int         `json:"consumer_version"`
	OwnerDomain                 string      `json:"owner_domain"`
	EventType                   string      `json:"event_type"`
	EventVersion                int         `json:"event_version"`
	QueueName                   string      `json:"queue_name"`
	EffectClass                 EffectClass `json:"effect_class"`
	DeliveryLimit               int         `json:"delivery_limit"`
	QuorumQueue                 bool        `json:"quorum_queue"`
	ManualAck                   bool        `json:"manual_ack"`
	InboxRetentionDays          int         `json:"inbox_retention_days"`
	DeadLetterFullPayloadDays   int         `json:"dead_letter_full_payload_days"`
	DeadLetterRedactedDays      int         `json:"dead_letter_redacted_days"`
	BrokerRedeliveryHorizonDays int         `json:"broker_redelivery_horizon_days"`
}

func NewConsumerPolicy(
	consumer string,
	consumerVersion int,
	domain string,
	eventType string,
	eventVersion int,
	queueName string,
	effectClass EffectClass,
	deliveryLimit int,
) (ConsumerPolicy, error) {
	value := ConsumerPolicy{
		ContractVersion: ConsumerPolicyContractVersion, Consumer: consumer, ConsumerVersion: consumerVersion,
		OwnerDomain: domain,
		EventType:   eventType, EventVersion: eventVersion, QueueName: queueName,
		EffectClass: effectClass, DeliveryLimit: deliveryLimit, QuorumQueue: true, ManualAck: true,
		InboxRetentionDays: InboxRetentionDays, DeadLetterFullPayloadDays: DeadLetterFullPayloadDays,
		DeadLetterRedactedDays:      DeadLetterRedactedRetentionDays,
		BrokerRedeliveryHorizonDays: BrokerRedeliveryHorizonDays,
	}
	if err := value.Validate(); err != nil {
		return ConsumerPolicy{}, err
	}
	return value, nil
}

func (value ConsumerPolicy) Validate() error {
	if value.ContractVersion != ConsumerPolicyContractVersion ||
		!validType(value.Consumer) || value.ConsumerVersion < 1 || !validDomain(value.OwnerDomain) ||
		!validType(value.EventType) || value.EventVersion < 1 ||
		!validType(value.QueueName) || !value.EffectClass.valid() {
		return fmt.Errorf("event consumer policy identity is invalid")
	}
	if value.DeliveryLimit < MinimumDeliveryLimit || value.DeliveryLimit > MaximumDeliveryLimit {
		return fmt.Errorf("event delivery limit must be between 1 and 25")
	}
	if !value.QuorumQueue || !value.ManualAck || value.InboxRetentionDays != InboxRetentionDays ||
		value.DeadLetterFullPayloadDays != DeadLetterFullPayloadDays ||
		value.DeadLetterRedactedDays != DeadLetterRedactedRetentionDays ||
		value.BrokerRedeliveryHorizonDays != BrokerRedeliveryHorizonDays {
		return fmt.Errorf("event reliability policy differs from ADR-0011")
	}
	return nil
}

func (value ConsumerPolicy) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire ConsumerPolicy
	return json.Marshal(wire(value))
}

func DecodeConsumerPolicy(document []byte) (ConsumerPolicy, error) {
	var value ConsumerPolicy
	if err := decodeStrict(document, &value); err != nil {
		return ConsumerPolicy{}, err
	}
	return value, value.Validate()
}

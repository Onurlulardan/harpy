package eventsv1

import (
	"encoding/json"
	"fmt"
)

type BootstrapReceipt struct {
	ContractVersion   string    `json:"contract_version"`
	ReceiptID         string    `json:"receipt_id"`
	Consumer          string    `json:"consumer"`
	ConsumerVersion   int       `json:"consumer_version"`
	TenantID          *string   `json:"tenant_id"`
	ProducerAuthority Authority `json:"producer_authority"`
	ConsumerAuthority Authority `json:"consumer_authority"`
	AggregateType     string    `json:"aggregate_type"`
	AggregateID       string    `json:"aggregate_id"`
	ThroughSequence   int64     `json:"through_sequence"`
	StateReference    string    `json:"state_reference"`
	StateSHA256       string    `json:"state_sha256"`
	CorrelationID     string    `json:"correlation_id"`
	CausationID       string    `json:"causation_id"`
	OccurredAt        string    `json:"occurred_at"`
	policy            ConsumerPolicy
}

func NewBootstrapReceipt(
	value BootstrapReceipt,
	consumers ConsumerCatalog,
) (BootstrapReceipt, error) {
	value.ContractVersion = BootstrapContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	if err := value.Validate(consumers); err != nil {
		return BootstrapReceipt{}, err
	}
	value.policy, _ = consumers.ResolveIdentity(value.Consumer, value.ConsumerVersion)
	return value, nil
}

func (value BootstrapReceipt) Validate(consumers ConsumerCatalog) error {
	if value.ContractVersion != BootstrapContractVersion || !validUUID(value.ReceiptID) ||
		!validType(value.Consumer) || value.ConsumerVersion < 1 {
		return fmt.Errorf("event bootstrap identity is invalid")
	}
	if err := value.ProducerAuthority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if err := value.ConsumerAuthority.ValidateConsumer(value.TenantID); err != nil {
		return err
	}
	consumer, consumerErr := consumers.ResolveIdentity(value.Consumer, value.ConsumerVersion)
	if consumerErr != nil || consumer.OwnerDomain != value.ConsumerAuthority.Domain {
		return fmt.Errorf("event bootstrap authority is not registered")
	}
	if !validType(value.AggregateType) || !validUUID(value.AggregateID) || value.ThroughSequence < 0 ||
		!validText(value.StateReference, MaximumReferenceBytes) || !lowerSHA256.MatchString(value.StateSHA256) {
		return fmt.Errorf("event bootstrap state is invalid")
	}
	if !validCorrelationID(value.CorrelationID) || !validUUID(value.CausationID) {
		return fmt.Errorf("event bootstrap audit identity is invalid")
	}
	if _, ok := validTimestamp(value.OccurredAt); !ok {
		return fmt.Errorf("event bootstrap timestamp is invalid")
	}
	return nil
}

func (value BootstrapReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(ConsumerCatalog{policies: []ConsumerPolicy{value.policy}}); err != nil {
		return nil, err
	}
	type wire BootstrapReceipt
	return json.Marshal(wire(value))
}

func DecodeBootstrapReceipt(
	document []byte,
	consumers ConsumerCatalog,
) (BootstrapReceipt, error) {
	var value BootstrapReceipt
	if err := decodeStrict(document, &value); err != nil {
		return BootstrapReceipt{}, err
	}
	if err := value.Validate(consumers); err != nil {
		return BootstrapReceipt{}, err
	}
	return NewBootstrapReceipt(value, consumers)
}

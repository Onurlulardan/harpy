package eventsv1

import (
	"encoding/json"
	"fmt"
)

type OperatorAction string

const (
	OperatorRetry         OperatorAction = "retry"
	OperatorRepairReissue OperatorAction = "repair_reissue"
)

type OperatorReceipt struct {
	ContractVersion         string         `json:"contract_version"`
	ReceiptID               string         `json:"receipt_id"`
	Action                  OperatorAction `json:"action"`
	Consumer                string         `json:"consumer"`
	ConsumerVersion         int            `json:"consumer_version"`
	TenantID                *string        `json:"tenant_id"`
	ProducerAuthority       Authority      `json:"producer_authority"`
	ConsumerAuthority       Authority      `json:"consumer_authority"`
	SourceEventID           string         `json:"source_event_id"`
	ResultEventID           string         `json:"result_event_id"`
	SourceBusinessEffectKey string         `json:"source_business_effect_idempotency_key"`
	ResultBusinessEffectKey string         `json:"result_business_effect_idempotency_key"`
	Reason                  string         `json:"reason"`
	TicketReference         string         `json:"ticket_reference"`
	ActorID                 string         `json:"actor_id"`
	CausationID             string         `json:"causation_id"`
	OccurredAt              string         `json:"occurred_at"`
	policy                  ConsumerPolicy
}

func NewOperatorReceipt(value OperatorReceipt, consumers ConsumerCatalog) (OperatorReceipt, error) {
	value.ContractVersion = OperatorReceiptContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	if err := value.Validate(consumers); err != nil {
		return OperatorReceipt{}, err
	}
	value.policy, _ = consumers.ResolveIdentity(value.Consumer, value.ConsumerVersion)
	return value, nil
}

func (value OperatorReceipt) Validate(consumers ConsumerCatalog) error {
	if value.ContractVersion != OperatorReceiptContractVersion || !validUUID(value.ReceiptID) ||
		!validType(value.Consumer) || value.ConsumerVersion < 1 ||
		!validUUID(value.SourceEventID) || !validUUID(value.ResultEventID) {
		return fmt.Errorf("event operator receipt identity is invalid")
	}
	if err := value.ProducerAuthority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if err := value.ConsumerAuthority.ValidateConsumer(value.TenantID); err != nil {
		return err
	}
	policy, err := consumers.ResolveIdentity(value.Consumer, value.ConsumerVersion)
	if err != nil || policy.OwnerDomain != value.ConsumerAuthority.Domain {
		return fmt.Errorf("event operator consumer authority is not registered")
	}
	if !validText(value.SourceBusinessEffectKey, MaximumKeyBytes) ||
		!validText(value.ResultBusinessEffectKey, MaximumKeyBytes) ||
		value.SourceBusinessEffectKey != value.ResultBusinessEffectKey ||
		!validText(value.Reason, MaximumReasonBytes) || !validText(value.TicketReference, MaximumTicketBytes) ||
		!validUUID(value.ActorID) || !validUUID(value.CausationID) {
		return fmt.Errorf("event operator receipt audit fields are invalid")
	}
	if _, ok := validTimestamp(value.OccurredAt); !ok {
		return fmt.Errorf("event operator DB-clock timestamp is invalid")
	}
	switch value.Action {
	case OperatorRetry:
		if value.ResultEventID != value.SourceEventID {
			return fmt.Errorf("event retry must preserve event ID")
		}
	case OperatorRepairReissue:
		if value.ResultEventID == value.SourceEventID {
			return fmt.Errorf("event repair must mint a new event ID")
		}
	default:
		return fmt.Errorf("event operator action is invalid")
	}
	return nil
}

func (value OperatorReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(ConsumerCatalog{policies: []ConsumerPolicy{value.policy}}); err != nil {
		return nil, err
	}
	type wire OperatorReceipt
	return json.Marshal(wire(value))
}

func DecodeOperatorReceipt(document []byte, consumers ConsumerCatalog) (OperatorReceipt, error) {
	var value OperatorReceipt
	if err := decodeStrict(document, &value); err != nil {
		return OperatorReceipt{}, err
	}
	if err := value.Validate(consumers); err != nil {
		return OperatorReceipt{}, err
	}
	return NewOperatorReceipt(value, consumers)
}

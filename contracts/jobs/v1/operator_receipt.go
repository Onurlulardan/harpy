package jobsv1

import (
	"encoding/json"
	"fmt"
)

type OperatorAction string

const (
	OperatorAdopt  OperatorAction = "adopt"
	OperatorCancel OperatorAction = "cancel"
	OperatorRetry  OperatorAction = "retry"
)

type OperatorReceipt struct {
	ContractVersion            string         `json:"contract_version"`
	ReceiptID                  string         `json:"receipt_id"`
	Action                     OperatorAction `json:"action"`
	TenantID                   *string        `json:"tenant_id"`
	Authority                  Authority      `json:"authority"`
	SourceJobID                *string        `json:"source_job_id"`
	SourceReference            *string        `json:"source_reference"`
	ResultJobID                *string        `json:"result_job_id"`
	SourceEffectIdempotencyKey *string        `json:"source_effect_idempotency_key"`
	ResultEffectIdempotencyKey *string        `json:"result_effect_idempotency_key"`
	Reason                     string         `json:"reason"`
	TicketReference            string         `json:"ticket_reference"`
	ActorID                    string         `json:"actor_id"`
	CausationID                string         `json:"causation_id"`
	OccurredAt                 string         `json:"occurred_at"`
}

func NewOperatorReceipt(value OperatorReceipt) (OperatorReceipt, error) {
	value.ContractVersion = OperatorReceiptContractVersion
	value.TenantID = cloneOptional(value.TenantID)
	value.SourceJobID = cloneOptional(value.SourceJobID)
	value.SourceReference = cloneOptional(value.SourceReference)
	value.ResultJobID = cloneOptional(value.ResultJobID)
	value.SourceEffectIdempotencyKey = cloneOptional(value.SourceEffectIdempotencyKey)
	value.ResultEffectIdempotencyKey = cloneOptional(value.ResultEffectIdempotencyKey)
	if err := value.Validate(); err != nil {
		return OperatorReceipt{}, err
	}
	return value, nil
}

func (value OperatorReceipt) Validate() error {
	if value.ContractVersion != OperatorReceiptContractVersion || !validUUID(value.ReceiptID) {
		return fmt.Errorf("operator receipt identity is invalid")
	}
	if err := value.Authority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if !validText(value.Reason, MaximumReasonBytes) ||
		!validText(value.TicketReference, MaximumTicketBytes) || !validUUID(value.ActorID) ||
		!validUUID(value.CausationID) {
		return fmt.Errorf("operator receipt audit fields are invalid")
	}
	if _, ok := validTimestamp(value.OccurredAt); !ok {
		return fmt.Errorf("operator receipt DB-clock timestamp is invalid")
	}
	switch value.Action {
	case OperatorAdopt:
		if value.SourceJobID != nil || value.SourceReference == nil ||
			!validText(*value.SourceReference, MaximumReferenceBytes) ||
			!validOptionalUUID(value.ResultJobID) || value.SourceEffectIdempotencyKey != nil ||
			value.ResultEffectIdempotencyKey != nil {
			return fmt.Errorf("adopt receipt fields are invalid")
		}
	case OperatorCancel:
		if !validOptionalUUID(value.SourceJobID) || value.SourceReference != nil || value.ResultJobID != nil ||
			value.SourceEffectIdempotencyKey != nil || value.ResultEffectIdempotencyKey != nil {
			return fmt.Errorf("cancel receipt fields are invalid")
		}
	case OperatorRetry:
		if !validOptionalUUID(value.SourceJobID) || value.SourceReference != nil ||
			!validOptionalUUID(value.ResultJobID) || *value.SourceJobID == *value.ResultJobID ||
			!validRetryEffectKeys(value.SourceEffectIdempotencyKey, value.ResultEffectIdempotencyKey) {
			return fmt.Errorf("retry receipt fields are invalid")
		}
	default:
		return fmt.Errorf("operator action is invalid")
	}
	return nil
}

func validOptionalUUID(value *string) bool { return value != nil && validUUID(*value) }

func validRetryEffectKeys(source *string, result *string) bool {
	if source == nil || result == nil {
		return source == nil && result == nil
	}
	return *source == *result && validText(*source, MaximumKeyBytes)
}

func (value OperatorReceipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire OperatorReceipt
	return json.Marshal(wire(value))
}

func DecodeOperatorReceipt(document []byte) (OperatorReceipt, error) {
	var value OperatorReceipt
	if err := decodeStrict(document, &value); err != nil {
		return OperatorReceipt{}, err
	}
	if err := value.Validate(); err != nil {
		return OperatorReceipt{}, err
	}
	return NewOperatorReceipt(value)
}

package effectsv1

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	ContractVersion  string    `json:"contract_version"`
	EffectID         string    `json:"effect_id"`
	EffectKind       string    `json:"effect_kind"`
	EffectRevision   int       `json:"effect_revision"`
	TenantID         *string   `json:"tenant_id"`
	Authority        Authority `json:"authority"`
	Provider         string    `json:"provider"`
	Operation        string    `json:"operation"`
	IdempotencyKey   string    `json:"idempotency_key"`
	CorrelationID    string    `json:"correlation_id"`
	CausationID      string    `json:"causation_id"`
	PayloadReference string    `json:"payload_reference"`
	PayloadSHA256    string    `json:"payload_sha256"`
	RequestedAt      string    `json:"requested_at"`
}

func NewRequest(
	effectID string,
	effectKind string,
	effectRevision int,
	tenantID *string,
	authority Authority,
	provider string,
	operation string,
	idempotencyKey string,
	correlationID string,
	causationID string,
	payloadReference string,
	payloadSHA256 string,
	requestedAt string,
) (Request, error) {
	value := Request{
		ContractVersion: RequestContractVersion,
		EffectID:        effectID, EffectKind: effectKind, EffectRevision: effectRevision,
		TenantID: cloneOptional(tenantID), Authority: authority,
		Provider: provider, Operation: operation, IdempotencyKey: idempotencyKey,
		CorrelationID: correlationID, CausationID: causationID,
		PayloadReference: payloadReference, PayloadSHA256: payloadSHA256,
		RequestedAt: requestedAt,
	}
	if err := value.Validate(); err != nil {
		return Request{}, err
	}
	return value, nil
}

func (value Request) Validate() error {
	if value.ContractVersion != RequestContractVersion || !validUUID(value.EffectID) {
		return fmt.Errorf("effect request contract identity is invalid")
	}
	if !effectKind.MatchString(value.EffectKind) || value.EffectRevision < 1 {
		return fmt.Errorf("effect request kind is invalid")
	}
	if err := value.Authority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if !providerName.MatchString(value.Provider) || !operationName.MatchString(value.Operation) ||
		!validBoundedText(value.IdempotencyKey, MaximumEffectIdempotencyKeyBytes) {
		return fmt.Errorf("effect request provider identity is invalid")
	}
	if !validCorrelationID(value.CorrelationID) || !validUUID(value.CausationID) {
		return fmt.Errorf("effect request causation identity is invalid")
	}
	if !validBoundedText(value.PayloadReference, MaximumPayloadReferenceBytes) ||
		!lowerSHA256.MatchString(value.PayloadSHA256) {
		return fmt.Errorf("effect request payload evidence is invalid")
	}
	if _, ok := validTimestamp(value.RequestedAt); !ok {
		return fmt.Errorf("effect request timestamp is invalid")
	}
	return nil
}

func (value Request) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire Request
	return json.Marshal(wire(value))
}

func DecodeRequest(document []byte) (Request, error) {
	var value Request
	if err := decodeStrict(document, &value); err != nil {
		return Request{}, err
	}
	value.TenantID = cloneOptional(value.TenantID)
	return value, value.Validate()
}

func (value Request) TenantIDClone() *string {
	return cloneOptional(value.TenantID)
}

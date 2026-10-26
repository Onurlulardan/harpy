package effectsv1

import "encoding/json"

type Receipt struct {
	ContractVersion   string        `json:"contract_version"`
	ReceiptID         string        `json:"receipt_id"`
	EffectID          string        `json:"effect_id"`
	EffectKind        string        `json:"effect_kind"`
	EffectRevision    int           `json:"effect_revision"`
	TenantID          *string       `json:"tenant_id"`
	Authority         Authority     `json:"authority"`
	Provider          string        `json:"provider"`
	Operation         string        `json:"operation"`
	IdempotencyKey    string        `json:"idempotency_key"`
	CorrelationID     string        `json:"correlation_id"`
	CausationID       string        `json:"causation_id"`
	Status            ReceiptStatus `json:"status"`
	Code              string        `json:"code"`
	ErrorClass        *ErrorClass   `json:"error_class"`
	ProviderReference *string       `json:"provider_reference"`
	ResultReference   *string       `json:"result_reference"`
	ResultSHA256      *string       `json:"result_sha256"`
	CompletedAt       string        `json:"completed_at"`
}

func NewSucceededReceipt(
	request Request,
	receiptID string,
	code string,
	providerReference *string,
	resultReference string,
	resultSHA256 string,
	completedAt string,
) (Receipt, error) {
	value := receiptFromRequest(request, receiptID, code, providerReference, completedAt)
	value.Status = ReceiptSucceeded
	value.ResultReference = cloneOptional(&resultReference)
	value.ResultSHA256 = cloneOptional(&resultSHA256)
	if err := value.ValidateForRequest(request); err != nil {
		return Receipt{}, err
	}
	return value, nil
}

func NewFailedReceipt(
	request Request,
	receiptID string,
	code string,
	errorClass ErrorClass,
	providerReference *string,
	completedAt string,
) (Receipt, error) {
	value := receiptFromRequest(request, receiptID, code, providerReference, completedAt)
	value.Status = ReceiptFailed
	value.ErrorClass = cloneErrorClass(&errorClass)
	if err := value.ValidateForRequest(request); err != nil {
		return Receipt{}, err
	}
	return value, nil
}

func receiptFromRequest(
	request Request,
	receiptID string,
	code string,
	providerReference *string,
	completedAt string,
) Receipt {
	return Receipt{
		ContractVersion: ReceiptContractVersion,
		ReceiptID:       receiptID, EffectID: request.EffectID,
		EffectKind: request.EffectKind, EffectRevision: request.EffectRevision,
		TenantID: cloneOptional(request.TenantID), Authority: request.Authority,
		Provider: request.Provider, Operation: request.Operation,
		IdempotencyKey: request.IdempotencyKey, Code: code,
		CorrelationID: request.CorrelationID, CausationID: request.CausationID,
		ProviderReference: cloneOptional(providerReference), CompletedAt: completedAt,
	}
}

func (value Receipt) MarshalJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	type wire Receipt
	return json.Marshal(wire(value))
}

func DecodeReceipt(document []byte) (Receipt, error) {
	var value Receipt
	if err := decodeStrict(document, &value); err != nil {
		return Receipt{}, err
	}
	value.TenantID = cloneOptional(value.TenantID)
	value.ErrorClass = cloneErrorClass(value.ErrorClass)
	value.ProviderReference = cloneOptional(value.ProviderReference)
	value.ResultReference = cloneOptional(value.ResultReference)
	value.ResultSHA256 = cloneOptional(value.ResultSHA256)
	return value, value.Validate()
}

func cloneErrorClass(value *ErrorClass) *ErrorClass {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

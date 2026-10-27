package postgres

import (
	"fmt"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/google/uuid"
)

type externalReceiptValues struct {
	receiptID         uuid.UUID
	tenantID          any
	requestSHA256     string
	effectID          uuid.UUID
	causationID       uuid.UUID
	completedAt       time.Time
	errorClass        *string
	providerReference *string
	resultReference   *string
	resultSHA256      *string
}

func validateExternalReceipt(
	permit ExternalEffectPermit,
	receipt effectsv1.Receipt,
) (externalReceiptValues, error) {
	if err := receipt.ValidateForRequest(permit.request); err != nil {
		return externalReceiptValues{}, fmt.Errorf("validate effect receipt: %w", err)
	}
	receiptID, err := uuid.Parse(receipt.ReceiptID)
	if err != nil {
		return externalReceiptValues{}, ErrExternalEffectReceipt
	}
	effectID, err := uuid.Parse(receipt.EffectID)
	if err != nil {
		return externalReceiptValues{}, ErrExternalEffectReceipt
	}
	causationID, err := uuid.Parse(receipt.CausationID)
	if err != nil {
		return externalReceiptValues{}, ErrExternalEffectReceipt
	}
	completedAt, err := time.Parse(time.RFC3339Nano, receipt.CompletedAt)
	if err != nil {
		return externalReceiptValues{}, ErrExternalEffectReceipt
	}
	requestDigest, err := externalRequestDigest(permit.request)
	if err != nil {
		return externalReceiptValues{}, ErrExternalEffectReceipt
	}
	values := externalReceiptValues{
		receiptID: receiptID, requestSHA256: requestDigest,
		effectID: effectID, causationID: causationID,
		completedAt: completedAt, providerReference: receipt.ProviderReference,
		resultReference: receipt.ResultReference, resultSHA256: receipt.ResultSHA256,
	}
	if permit.claim.tenantScoped {
		values.tenantID = permit.claim.tenantID
	}
	if receipt.ErrorClass != nil {
		value := string(*receipt.ErrorClass)
		values.errorClass = &value
	}
	return values, nil
}

func externalReceiptArguments(
	permit ExternalEffectPermit,
	receipt effectsv1.Receipt,
	values externalReceiptValues,
) []any {
	claim := permit.claim
	return []any{
		values.receiptID, values.tenantID,
		claim.authority.scopeKind, claim.authority.scopeID,
		claim.authority.groupID, claim.authority.groupGeneration,
		claim.authority.domain, claim.authority.generation,
		values.requestSHA256, values.effectID, receipt.EffectKind,
		receipt.EffectRevision, receipt.IdempotencyKey, receipt.Provider,
		receipt.Operation, receipt.CorrelationID, values.causationID,
		string(receipt.Status), values.errorClass, receipt.Code,
		values.providerReference, values.resultReference, values.resultSHA256,
		values.completedAt,
	}
}

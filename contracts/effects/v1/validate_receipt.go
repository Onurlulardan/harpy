package effectsv1

import "fmt"

func (value Receipt) Validate() error {
	if value.ContractVersion != ReceiptContractVersion || !validUUID(value.ReceiptID) ||
		!validUUID(value.EffectID) {
		return fmt.Errorf("effect receipt contract identity is invalid")
	}
	if !effectKind.MatchString(value.EffectKind) || value.EffectRevision < 1 ||
		!value.Status.valid() {
		return fmt.Errorf("effect receipt kind or status is invalid")
	}
	if err := value.Authority.ValidateTenant(value.TenantID); err != nil {
		return err
	}
	if !providerName.MatchString(value.Provider) || !operationName.MatchString(value.Operation) ||
		!validBoundedText(value.IdempotencyKey, MaximumEffectIdempotencyKeyBytes) {
		return fmt.Errorf("effect receipt provider identity is invalid")
	}
	if !validCorrelationID(value.CorrelationID) || !validUUID(value.CausationID) {
		return fmt.Errorf("effect receipt causation identity is invalid")
	}
	if !machineCode.MatchString(value.Code) || value.Code == JobOutcomeUnknownCode {
		return fmt.Errorf("effect receipt code is invalid")
	}
	if value.ProviderReference != nil &&
		!validBoundedText(*value.ProviderReference, MaximumProviderReferenceBytes) {
		return fmt.Errorf("effect receipt provider reference is invalid")
	}
	if _, ok := validTimestamp(value.CompletedAt); !ok {
		return fmt.Errorf("effect receipt timestamp is invalid")
	}
	return value.validateOutcome()
}

func (value Receipt) validateOutcome() error {
	switch value.Status {
	case ReceiptSucceeded:
		if value.ErrorClass != nil || value.ResultReference == nil || value.ResultSHA256 == nil {
			return fmt.Errorf("succeeded receipt nullable fields are invalid")
		}
		if !validBoundedText(*value.ResultReference, MaximumResultReferenceBytes) ||
			!lowerSHA256.MatchString(*value.ResultSHA256) {
			return fmt.Errorf("succeeded receipt result evidence is invalid")
		}
	case ReceiptFailed:
		if value.ErrorClass == nil || !value.ErrorClass.valid() ||
			value.ResultReference != nil || value.ResultSHA256 != nil {
			return fmt.Errorf("failed receipt nullable fields are invalid")
		}
	default:
		return fmt.Errorf("effect receipt status is invalid")
	}
	return nil
}

func (value Receipt) ValidateForRequest(request Request) error {
	if err := request.Validate(); err != nil {
		return fmt.Errorf("effect request: %w", err)
	}
	if err := value.Validate(); err != nil {
		return err
	}
	if value.EffectID != request.EffectID || value.EffectKind != request.EffectKind ||
		value.EffectRevision != request.EffectRevision || !equalOptional(value.TenantID, request.TenantID) ||
		value.Authority != request.Authority || value.Provider != request.Provider ||
		value.Operation != request.Operation || value.IdempotencyKey != request.IdempotencyKey ||
		value.CorrelationID != request.CorrelationID || value.CausationID != request.CausationID {
		return fmt.Errorf("effect receipt does not match request")
	}
	requestedAt, _ := validTimestamp(request.RequestedAt)
	completedAt, _ := validTimestamp(value.CompletedAt)
	if completedAt.Before(requestedAt) {
		return fmt.Errorf("effect receipt predates request")
	}
	return nil
}

func (value Receipt) TenantIDClone() *string {
	return cloneOptional(value.TenantID)
}

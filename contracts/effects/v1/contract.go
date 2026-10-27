package effectsv1

const (
	RequestContractVersion           = "effect-request.v1"
	ReceiptContractVersion           = "effect-receipt.v1"
	JobOutcomeUnknownCode            = "JOB_OUTCOME_UNKNOWN"
	MaximumDomainBytes               = 63
	MaximumEffectKindBytes           = 128
	MaximumOperationBytes            = 128
	MaximumProviderBytes             = 128
	MaximumReceiptCodeBytes          = 64
	MaximumCorrelationIDBytes        = 64
	MaximumEffectIdempotencyKeyBytes = 512
	MaximumProviderReferenceBytes    = 512
	MaximumPayloadReferenceBytes     = 2048
	MaximumResultReferenceBytes      = 2048
	MaximumDocumentBytes             = 131072
)

type ErrorClass string

const (
	ErrorTransient ErrorClass = "transient"
	ErrorPermanent ErrorClass = "permanent"
	ErrorConflict  ErrorClass = "conflict"
	ErrorCancelled ErrorClass = "cancelled"
)

func (value ErrorClass) valid() bool {
	switch value {
	case ErrorTransient, ErrorPermanent, ErrorConflict, ErrorCancelled:
		return true
	default:
		return false
	}
}

type ReceiptStatus string

const (
	ReceiptSucceeded ReceiptStatus = "succeeded"
	ReceiptFailed    ReceiptStatus = "failed"
)

func (value ReceiptStatus) valid() bool {
	return value == ReceiptSucceeded || value == ReceiptFailed
}

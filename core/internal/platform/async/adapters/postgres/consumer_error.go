package postgres

import "errors"

var (
	ErrConsumerBindingInvalid       = errors.New("async consumer binding is invalid")
	ErrConsumerAuthorityStale       = errors.New("async consumer authority is stale")
	ErrConsumerInputInvalid         = errors.New("async consumer input is invalid")
	ErrProducerEvidenceMissing      = errors.New("async producer evidence is missing")
	ErrEventGapConflict             = errors.New("async event sequence gap evidence conflicts")
	ErrEventPermitInvalid           = errors.New("async event effect permit is invalid")
	ErrEventReceiptInvalid          = errors.New("async event effect receipt is invalid")
	ErrExternalEffectReceiptMissing = errors.New("external effect receipt is not durable")
	ErrBootstrapPermitInvalid       = errors.New("async bootstrap permit is invalid")
	ErrBootstrapStreamNotEmpty      = errors.New("async bootstrap requires an empty stream")
)

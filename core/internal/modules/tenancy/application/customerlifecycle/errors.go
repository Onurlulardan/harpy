package customerlifecycle

import "errors"

var (
	ErrAuthentication      = errors.New("customer lifecycle authentication failed")
	ErrForbidden           = errors.New("customer lifecycle operation is forbidden")
	ErrUnavailable         = errors.New("customer lifecycle operation is unavailable")
	ErrModeDenied          = errors.New("customer lifecycle operation requires SaaS mode")
	ErrInvalid             = errors.New("customer lifecycle request is invalid")
	ErrNotFound            = errors.New("customer lifecycle resource was not found")
	ErrRevisionConflict    = errors.New("customer lifecycle revision conflict")
	ErrIdempotencyConflict = errors.New("customer lifecycle idempotency conflict")
	ErrStateConflict       = errors.New("customer lifecycle state conflict")
	ErrPurgeProof          = errors.New("customer lifecycle purge proof is unavailable")
)

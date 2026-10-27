package billingapi

import "errors"

var (
	ErrMeteredResourceConfiguration     = errors.New("billing metered resource configuration is invalid")
	ErrMeteredResourceRequest           = errors.New("billing metered resource request is invalid")
	ErrMeteredResourceUnavailable       = errors.New("billing metered resource dependency is unavailable")
	ErrMeteredResourceIdempotencyReused = errors.New(
		"billing metered resource idempotency key was reused",
	)
	ErrMeteredResourceStateConflict = errors.New("billing metered resource state conflicts")
	ErrMeteredResourceEvidence      = errors.New("billing metered resource evidence is invalid")
)

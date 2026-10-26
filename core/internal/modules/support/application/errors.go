package application

import supportports "github.com/Onurlulardan/harpy/core/internal/modules/support/ports"

var (
	ErrAuthentication      = supportports.ErrAuthentication
	ErrForbidden           = supportports.ErrForbidden
	ErrInvalid             = supportports.ErrInvalid
	ErrNotFound            = supportports.ErrNotFound
	ErrConflict            = supportports.ErrConflict
	ErrRateLimited         = supportports.ErrRateLimited
	ErrUnavailable         = supportports.ErrUnavailable
	ErrIdempotencyMismatch = supportports.ErrIdempotencyMismatch
	ErrRevisionStale       = supportports.ErrRevisionStale
)

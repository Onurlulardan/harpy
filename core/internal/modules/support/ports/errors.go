package ports

import "errors"

var (
	ErrAuthentication      = errors.New("support authentication failed")
	ErrForbidden           = errors.New("support operation is forbidden")
	ErrInvalid             = errors.New("support input is invalid")
	ErrNotFound            = errors.New("support resource was not found")
	ErrConflict            = errors.New("support state conflicts")
	ErrRateLimited         = errors.New("support public rate limit exceeded")
	ErrUnavailable         = errors.New("support authority is unavailable")
	ErrIdempotencyMismatch = errors.New("support idempotency key was reused")
	ErrRevisionStale       = errors.New("support revision is stale")
)

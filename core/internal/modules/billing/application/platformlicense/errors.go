package platformlicense

import "errors"

var (
	ErrInvalid           = errors.New("billing platform license request is invalid")
	ErrNotFound          = errors.New("billing platform license resource was not found")
	ErrIdempotencyReused = errors.New(
		"billing platform license idempotency key was reused",
	)
	ErrConflict    = errors.New("billing platform license request conflicts")
	ErrInProgress  = errors.New("billing platform license refresh is in progress")
	ErrForbidden   = errors.New("billing platform license request is forbidden")
	ErrUnavailable = errors.New("billing platform license authority is unavailable")
)

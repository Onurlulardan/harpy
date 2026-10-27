package dashboardexport

import "errors"

var (
	ErrInvalid           = errors.New("analytics dashboard export input is invalid")
	ErrForbidden         = errors.New("analytics dashboard export is forbidden")
	ErrNotFound          = errors.New("analytics dashboard export was not found")
	ErrConflict          = errors.New("analytics dashboard export conflicts")
	ErrIdempotencyReused = errors.New("analytics dashboard export idempotency key was reused")
	ErrQueueLimit        = errors.New("analytics dashboard export admission is full")
	ErrNotReady          = errors.New("analytics dashboard export is not ready")
	ErrExpired           = errors.New("analytics dashboard export expired")
	ErrOutcomeUnknown    = errors.New("analytics dashboard export outcome is unknown")
	ErrUnavailable       = errors.New("analytics dashboard export is unavailable")
)

package ports

import "errors"

var (
	ErrNotFound             = errors.New("platform authority record not found")
	ErrConflict             = errors.New("platform authority record conflicts")
	ErrStaleHead            = errors.New("platform authority head is stale")
	ErrFinalizerUnavailable = errors.New("platform authority owner finalizer is unavailable")
)

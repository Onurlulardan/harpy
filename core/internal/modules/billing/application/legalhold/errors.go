package legalhold

import "errors"

var (
	ErrConfiguration = errors.New("billing legal hold configuration is invalid")
	ErrRequest       = errors.New("billing legal hold request is invalid")
	ErrResult        = errors.New("billing legal hold result is invalid")
	ErrConflict      = errors.New("billing legal hold idempotency conflict")
	ErrNotFound      = errors.New("billing legal hold was not found")
	ErrDependency    = errors.New("billing legal hold authority is not ready")
)

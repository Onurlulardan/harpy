package api

import "errors"

var (
	ErrInvalid     = errors.New("platform operations request is invalid")
	ErrDenied      = errors.New("platform operations request is denied")
	ErrConflict    = errors.New("platform operations state conflicts")
	ErrNotFound    = errors.New("platform operations resource was not found")
	ErrUnavailable = errors.New("platform operations dependency is unavailable")
)

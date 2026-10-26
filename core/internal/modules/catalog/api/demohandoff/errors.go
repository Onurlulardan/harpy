package demohandoff

import "errors"

var (
	ErrInvalid     = errors.New("Catalog demo handoff is invalid")
	ErrConflict    = errors.New("Catalog demo handoff conflicts")
	ErrUnavailable = errors.New("Catalog demo handoff is unavailable")
)

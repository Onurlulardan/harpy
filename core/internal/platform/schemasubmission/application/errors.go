package application

import "errors"

var (
	ErrInvalid     = errors.New("schema submission is invalid")
	ErrConflict    = errors.New("schema submission conflicts with persisted state")
	ErrUnavailable = errors.New("schema submission is unavailable")
)

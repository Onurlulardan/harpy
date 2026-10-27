package api

import "errors"

var (
	ErrInvalid     = errors.New("system resource layout is invalid")
	ErrNotFound    = errors.New("system resource layout was not found")
	ErrConflict    = errors.New("system resource layout revision conflicts")
	ErrUnavailable = errors.New("system resource layout storage is unavailable")
)

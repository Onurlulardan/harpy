package application

import "errors"

var (
	ErrConfiguration = errors.New("notification application configuration is invalid")
	ErrInvalid       = errors.New("notification operation input is invalid")
	ErrForbidden     = errors.New("notification operation is forbidden")
	ErrNotFound      = errors.New("notification resource was not found")
	ErrConflict      = errors.New("notification operation conflicts")
	ErrUnavailable   = errors.New("notification dependency is unavailable")
)

package api

import "errors"

var (
	ErrInvalid = errors.New("tenant activation capability is invalid")
	ErrClosed  = errors.New("tenant activation capability is closed")
	ErrDenied  = errors.New("tenant activation command is denied")
	ErrOrder   = errors.New("tenant activation command order is invalid")
)

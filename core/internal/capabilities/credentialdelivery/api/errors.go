package api

import "errors"

var (
	ErrInvalid = errors.New("credential delivery capability is invalid")
	ErrClosed  = errors.New("credential delivery capability is closed")
	ErrDenied  = errors.New("credential delivery command is denied")
	ErrOrder   = errors.New("credential delivery command order is invalid")
)

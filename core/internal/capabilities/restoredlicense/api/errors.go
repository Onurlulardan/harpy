package api

import "errors"

var (
	ErrInvalid = errors.New("restored license capability is invalid")
	ErrClosed  = errors.New("restored license capability is closed")
	ErrDenied  = errors.New("restored license command is denied")
	ErrOrder   = errors.New("restored license command order is invalid")
)

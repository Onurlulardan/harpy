package api

import "errors"

var (
	ErrInvalid = errors.New("license bootstrap capability is invalid")
	ErrClosed  = errors.New("license bootstrap capability is closed")
	ErrDenied  = errors.New("license bootstrap command is denied")
	ErrOrder   = errors.New("license bootstrap command order is invalid")
)

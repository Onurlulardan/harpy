package adoption

import "errors"

var (
	ErrInvalid     = errors.New("support adoption input is invalid")
	ErrSourceDrift = errors.New("support adoption source changed")
	ErrConflict    = errors.New("support adoption state conflicts")
	ErrLease       = errors.New("support adoption lease is unavailable")
	ErrUnavailable = errors.New("support adoption authority is unavailable")
)

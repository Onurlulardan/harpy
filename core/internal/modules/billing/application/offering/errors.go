package offering

import "errors"

var (
	ErrConfiguration = errors.New("billing offering commands are misconfigured")
	ErrRequest       = errors.New("billing offering request is invalid")
	ErrResult        = errors.New("billing offering result is invalid")
)

package period

import "errors"

var (
	ErrCadence = errors.New("billing period cadence is invalid")
	ErrInstant = errors.New("billing period instant is invalid")
	ErrRange   = errors.New("billing period range is invalid")
	ErrAnchor  = errors.New("billing period anchor is invalid")
)

package usage

import "errors"

var (
	ErrAdmission  = errors.New("billing usage admission is invalid")
	ErrGauge      = errors.New("billing usage gauge adjustment is invalid")
	ErrTransition = errors.New("billing usage reservation transition is invalid")
	ErrExpiry     = errors.New("billing usage reservation expiry evidence is invalid")
)

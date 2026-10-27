package usage

import "errors"

var (
	ErrConfiguration     = errors.New("billing usage commands are misconfigured")
	ErrRequest           = errors.New("billing usage request is invalid")
	ErrSnapshot          = errors.New("billing usage snapshot is invalid")
	ErrReservationReused = errors.New("billing usage reservation operation was reused")
	ErrOperationReused   = errors.New("billing usage owner operation was reused")
	ErrReservationAbsent = errors.New("billing usage reservation does not exist")
	ErrState             = errors.New("billing usage reservation state conflicts")
)

package discount

import "errors"

var (
	ErrConfiguration   = errors.New("billing discount commands are misconfigured")
	ErrRequest         = errors.New("billing discount redemption request is invalid")
	ErrSnapshot        = errors.New("billing discount eligibility snapshot is invalid")
	ErrIneligible      = errors.New("billing discount is not eligible")
	ErrOperationReused = errors.New("billing discount operation was reused")
)

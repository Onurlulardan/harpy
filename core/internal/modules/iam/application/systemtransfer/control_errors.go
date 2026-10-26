package systemtransfer

import "errors"

var (
	ErrControlInvalid     = errors.New("IAM transfer control input is invalid")
	ErrControlUnavailable = errors.New("IAM transfer control store is unavailable")
	ErrControlNotFound    = errors.New("IAM transfer control row was not found")
	ErrControlConflict    = errors.New("IAM transfer control row conflicts")
	ErrControlCapacity    = errors.New("IAM transfer control capacity is exhausted")
	ErrControlExpired     = errors.New("IAM transfer artifact expired")
)

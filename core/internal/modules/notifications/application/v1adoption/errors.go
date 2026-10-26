package v1adoption

import "errors"

var (
	ErrInvalid     = errors.New("notification V1 adoption input is invalid")
	ErrConflict    = errors.New("notification V1 adoption state conflicts")
	ErrBlocked     = errors.New("notification V1 adoption is blocked")
	ErrUnavailable = errors.New("notification V1 adoption storage is unavailable")
)

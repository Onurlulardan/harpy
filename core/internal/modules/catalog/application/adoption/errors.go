package adoption

import "errors"

var (
	ErrInvalid     = errors.New("catalog adoption input is invalid")
	ErrForbidden   = errors.New("catalog adoption operation is forbidden")
	ErrBlocked     = errors.New("catalog adoption is blocked")
	ErrConflict    = errors.New("catalog adoption operation conflicts")
	ErrUnavailable = errors.New("catalog adoption authority is unavailable")
)

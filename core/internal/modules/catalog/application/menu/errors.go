package menu

import "errors"

var (
	ErrInvalid     = errors.New("catalog menu input is invalid")
	ErrForbidden   = errors.New("catalog menu operation is forbidden")
	ErrNotFound    = errors.New("catalog menu was not found")
	ErrConflict    = errors.New("catalog menu revision conflicts")
	ErrUnavailable = errors.New("catalog menu authority is unavailable")
)

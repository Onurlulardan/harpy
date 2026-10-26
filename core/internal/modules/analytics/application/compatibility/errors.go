package compatibility

import "errors"

var (
	ErrInvalid     = errors.New("analytics compatibility input is invalid")
	ErrNotFound    = errors.New("analytics legacy mapping was not found")
	ErrConflict    = errors.New("analytics compatibility state conflicts")
	ErrBlocked     = errors.New("analytics compatibility retirement is blocked")
	ErrUnavailable = errors.New("analytics compatibility authority is unavailable")
)

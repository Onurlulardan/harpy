package deletion

import "errors"

var (
	ErrInvalid     = errors.New("documents deletion input is invalid")
	ErrConflict    = errors.New("documents deletion lease conflicts")
	ErrNotFound    = errors.New("documents deletion was not found")
	ErrUnavailable = errors.New("documents deletion dependency is unavailable")
)

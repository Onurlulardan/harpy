package folder

import "errors"

var (
	ErrInvalid     = errors.New("documents folder input is invalid")
	ErrForbidden   = errors.New("documents folder operation is forbidden")
	ErrNotFound    = errors.New("documents folder was not found")
	ErrConflict    = errors.New("documents folder revision conflicts")
	ErrUnavailable = errors.New("documents folder storage is unavailable")
)

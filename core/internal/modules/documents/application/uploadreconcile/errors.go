package uploadreconcile

import "errors"

var (
	ErrInvalid     = errors.New("documents upload reconciliation input is invalid")
	ErrConflict    = errors.New("documents upload reconciliation state conflicts")
	ErrUnavailable = errors.New("documents upload reconciliation dependency is unavailable")
)

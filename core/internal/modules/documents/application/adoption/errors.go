package adoption

import "errors"

var (
	ErrInvalid     = errors.New("documents adoption input is invalid")
	ErrConflict    = errors.New("documents adoption state conflicts")
	ErrBlocked     = errors.New("documents adoption reconciliation is blocked")
	ErrUnavailable = errors.New("documents adoption dependency is unavailable")
)

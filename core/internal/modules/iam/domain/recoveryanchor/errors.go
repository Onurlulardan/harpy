package recoveryanchor

import "errors"

var (
	ErrInvalid    = errors.New("recovery anchor request is invalid")
	ErrDenied     = errors.New("recovery anchor operation is denied")
	ErrNotFound   = errors.New("recovery anchor was not found")
	ErrConflict   = errors.New("recovery anchor state conflicts")
	ErrDependency = errors.New("recovery anchor dependency is unavailable")
)

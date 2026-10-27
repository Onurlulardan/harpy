package download

import "errors"

var (
	ErrInvalid     = errors.New("documents download input is invalid")
	ErrForbidden   = errors.New("documents download is forbidden")
	ErrNotFound    = errors.New("documents version was not found")
	ErrIntegrity   = errors.New("documents content integrity failed")
	ErrUnavailable = errors.New("documents download dependency is unavailable")
)

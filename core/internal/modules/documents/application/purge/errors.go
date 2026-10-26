package purge

import "errors"

var (
	ErrInvalid     = errors.New("documents purge input is invalid")
	ErrNotFound    = errors.New("documents purge candidate was not found")
	ErrLegalHold   = errors.New("documents purge is blocked by a legal hold")
	ErrReference   = errors.New("documents purge is blocked by a lifecycle reference")
	ErrConflict    = errors.New("documents purge state conflicts")
	ErrUnavailable = errors.New("documents purge dependency is unavailable")
)

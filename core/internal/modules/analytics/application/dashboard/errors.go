package dashboard

import "errors"

var (
	ErrInvalid     = errors.New("analytics dashboard request is invalid")
	ErrForbidden   = errors.New("analytics dashboard operation is forbidden")
	ErrNotFound    = errors.New("analytics dashboard was not found")
	ErrConflict    = errors.New("analytics dashboard conflicts with current state")
	ErrEntitlement = errors.New("analytics dashboard entitlement denied the operation")
	ErrUnavailable = errors.New("analytics dashboard dependency is unavailable")
)

package publication

import "errors"

var (
	ErrInvalid              = errors.New("analytics publication request is invalid")
	ErrForbidden            = errors.New("analytics publication operation is forbidden")
	ErrNotFound             = errors.New("analytics publication was not found")
	ErrConflict             = errors.New("analytics publication conflicts with current state")
	ErrEntitlement          = errors.New("analytics publication entitlement denied the operation")
	ErrRateLimitUnavailable = errors.New("analytics public rate limit is unavailable")
	ErrRateLimited          = errors.New("analytics public rate limit exceeded")
	ErrUnavailable          = errors.New("analytics publication dependency is unavailable")
)

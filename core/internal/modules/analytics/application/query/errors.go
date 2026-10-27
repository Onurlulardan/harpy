package query

import "errors"

var (
	ErrInvalid     = errors.New("analytics query request is invalid")
	ErrForbidden   = errors.New("analytics query operation is forbidden")
	ErrNotFound    = errors.New("analytics query was not found")
	ErrConflict    = errors.New("analytics query conflicts with current state")
	ErrEntitlement = errors.New("analytics query entitlement denied the operation")
	ErrOwnerNotV2  = errors.New("analytics query owner is not V2")
	ErrUnavailable = errors.New("analytics query dependency is unavailable")
)

package systemresource

import "errors"

var (
	ErrContract    = errors.New("IAM system-resource contract is invalid")
	ErrForbidden   = errors.New("IAM system-resource operation is forbidden")
	ErrNotFound    = errors.New("IAM system-resource record was not found")
	ErrConflict    = errors.New("IAM system-resource operation conflicts")
	ErrUnavailable = errors.New("IAM system-resource operation is unavailable")
)

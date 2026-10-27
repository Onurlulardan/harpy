package systemresources

import "errors"

var (
	ErrContract    = errors.New("system-resource facade contract is invalid")
	ErrForbidden   = errors.New("system-resource facade operation is forbidden")
	ErrNotFound    = errors.New("system-resource facade record was not found")
	ErrConflict    = errors.New("system-resource facade operation conflicts")
	ErrUnavailable = errors.New("system-resource facade is unavailable")
)

package item

import "errors"

var (
	ErrInvalid     = errors.New("documents item input is invalid")
	ErrForbidden   = errors.New("documents item operation is forbidden")
	ErrNotFound    = errors.New("documents item was not found")
	ErrConflict    = errors.New("documents item revision conflicts")
	ErrLegalHold   = errors.New("documents item is under legal hold")
	ErrUnavailable = errors.New("documents item dependency is unavailable")
)

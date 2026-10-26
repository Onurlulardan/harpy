package legalhold

import "errors"

var (
	ErrInvalid     = errors.New("documents legal hold input is invalid")
	ErrForbidden   = errors.New("documents legal hold source is forbidden")
	ErrNotFound    = errors.New("documents legal hold was not found")
	ErrConflict    = errors.New("documents legal hold target conflicts")
	ErrUnavailable = errors.New("documents legal hold dependency is unavailable")
)

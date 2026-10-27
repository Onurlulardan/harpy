package knowledgefile

import "errors"

var (
	ErrInvalid     = errors.New("Documents knowledge file request is invalid")
	ErrForbidden   = errors.New("Documents knowledge file access is forbidden")
	ErrConflict    = errors.New("Documents knowledge file state conflicts")
	ErrUnavailable = errors.New("Documents knowledge file owner is unavailable")
)

package formrecordarchive

import "errors"

var (
	ErrInvalid           = errors.New("Documents form-record archive request is invalid")
	ErrForbidden         = errors.New("Documents form-record archive access is forbidden")
	ErrReferenceMismatch = errors.New("Documents form-record archive links do not match")
	ErrNotFound          = errors.New("Documents form-record archive was not found")
	ErrConflict          = errors.New("Documents form-record archive state conflicts")
	ErrUnavailable       = errors.New("Documents form-record archive owner is unavailable")
)

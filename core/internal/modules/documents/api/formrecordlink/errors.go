package formrecordlink

import "errors"

var (
	ErrInvalid          = errors.New("documents form-record link request is invalid")
	ErrReferenceInvalid = errors.New("documents form-record reference is unavailable")
	ErrUnavailable      = errors.New("documents form-record link registry is unavailable")
)

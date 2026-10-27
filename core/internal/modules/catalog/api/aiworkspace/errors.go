package aiworkspace

import "errors"

var (
	ErrInvalid     = errors.New("Catalog AI workspace request is invalid")
	ErrNotFound    = errors.New("Catalog AI workspace item was not found")
	ErrUnavailable = errors.New("Catalog AI workspace owner is unavailable")
)

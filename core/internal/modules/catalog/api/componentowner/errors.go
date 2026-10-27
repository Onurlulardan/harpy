package componentowner

import "errors"

var (
	ErrInvalid     = errors.New("catalog component owner command is invalid")
	ErrUnavailable = errors.New("catalog component owner is unavailable")
)

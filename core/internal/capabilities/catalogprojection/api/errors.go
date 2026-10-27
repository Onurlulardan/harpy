package catalogprojection

import "errors"

var (
	ErrInvalid     = errors.New("catalog materialization request is invalid")
	ErrConflict    = errors.New("catalog materialization conflicts")
	ErrUnavailable = errors.New("catalog materialization is unavailable")
)

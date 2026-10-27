package demomaterialization

import "errors"

var (
	ErrInvalid     = errors.New("catalog demo materialization request is invalid")
	ErrConflict    = errors.New("catalog demo materialization conflicts")
	ErrUnavailable = errors.New("catalog demo materialization is unavailable")
)

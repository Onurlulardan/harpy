package installationdelivery

import "errors"

var (
	ErrInvalid     = errors.New("installation credential delivery is invalid")
	ErrUnavailable = errors.New("installation credential delivery is unavailable")
	ErrConflict    = errors.New("installation credential delivery changed")
)

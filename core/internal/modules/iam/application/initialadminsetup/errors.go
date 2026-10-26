package initialadminsetup

import "errors"

var (
	ErrInvalid     = errors.New("initial admin setup is invalid")
	ErrUnavailable = errors.New("initial admin setup is unavailable")
)

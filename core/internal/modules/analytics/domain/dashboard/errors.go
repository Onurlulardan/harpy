package dashboard

import "errors"

var (
	ErrInvalid  = errors.New("analytics dashboard is invalid")
	ErrConflict = errors.New("analytics dashboard conflicts with current state")
)

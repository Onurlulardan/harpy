package publication

import "errors"

var (
	ErrInvalid  = errors.New("analytics publication is invalid")
	ErrConflict = errors.New("analytics publication conflicts with current state")
)

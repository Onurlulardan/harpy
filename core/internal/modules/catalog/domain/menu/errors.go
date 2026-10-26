package menu

import "errors"

var (
	ErrInvalid  = errors.New("catalog menu is invalid")
	ErrConflict = errors.New("catalog menu revision conflicts")
)

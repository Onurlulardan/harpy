package postgres

import "errors"

var (
	ErrInvalid  = errors.New("notification postgres operation is invalid")
	ErrConflict = errors.New("notification postgres mutation conflicts")
	ErrNotFound = errors.New("notification postgres resource was not found")
)

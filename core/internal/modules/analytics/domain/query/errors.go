package query

import "errors"

var (
	ErrInvalid  = errors.New("analytics query is invalid")
	ErrConflict = errors.New("analytics query conflicts with current state")
)

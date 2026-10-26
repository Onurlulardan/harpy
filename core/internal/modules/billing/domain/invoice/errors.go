package invoice

import "errors"

var (
	ErrLine      = errors.New("billing invoice line is invalid")
	ErrInvariant = errors.New("billing invoice snapshot invariant failed")
	ErrTotals    = errors.New("billing invoice totals are invalid")
)

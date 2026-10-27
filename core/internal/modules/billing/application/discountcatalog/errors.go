package discountcatalog

import "errors"

var (
	ErrConfiguration = errors.New("billing discount catalog commands are misconfigured")
	ErrRequest       = errors.New("billing discount catalog request is invalid")
	ErrResult        = errors.New("billing discount catalog result is invalid")
)

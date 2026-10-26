package campaign

import "errors"

var (
	ErrConfiguration = errors.New("billing campaign commands are misconfigured")
	ErrRequest       = errors.New("billing campaign request is invalid")
	ErrResult        = errors.New("billing campaign result is invalid")
)

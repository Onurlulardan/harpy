package catalogquery

import "errors"

var (
	ErrConfiguration = errors.New("billing catalog queries are misconfigured")
	ErrRequest       = errors.New("billing catalog query is invalid")
	ErrNotFound      = errors.New("billing public offering was not found")
	ErrResult        = errors.New("billing catalog query result is invalid")
)

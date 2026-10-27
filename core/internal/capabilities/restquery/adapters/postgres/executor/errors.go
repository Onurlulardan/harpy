package executor

import "errors"

var (
	ErrContract      = errors.New("REST Query PostgreSQL executor contract is invalid")
	ErrExecution     = errors.New("REST Query PostgreSQL execution failed")
	ErrDirectMarshal = errors.New("REST Query execution result cannot be marshaled directly")
)

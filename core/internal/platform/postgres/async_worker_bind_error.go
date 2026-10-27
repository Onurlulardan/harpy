package postgres

import "errors"

var (
	ErrAsyncWorkerProfile        = errors.New("postgres async store requires worker profile")
	ErrAsyncWorkerBindingInvalid = errors.New("postgres async worker binding is invalid")
)

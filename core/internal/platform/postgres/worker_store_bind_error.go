package postgres

import "errors"

var (
	ErrWorkerStoreProfile        = errors.New("postgres worker store requires worker profile")
	ErrWorkerStoreBindingInvalid = errors.New("postgres worker store binding is invalid")
)

package postgres

import "errors"

var (
	ErrAPIStoreProfile        = errors.New("postgres API store requires API profile")
	ErrAPIStoreBindingInvalid = errors.New("postgres API store binding is invalid")
)

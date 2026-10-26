package api

import "errors"

var (
	ErrConflict        = errors.New("protected authority operation conflicts")
	ErrStaleHead       = errors.New("protected authority head is stale")
	ErrNotReady        = errors.New("protected authority operation is not ready")
	ErrBindingConflict = errors.New("protected authority binding conflicts")
	ErrDependency      = errors.New("protected authority dependency failed")
)

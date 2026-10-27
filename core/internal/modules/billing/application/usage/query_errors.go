package usage

import "errors"

var (
	ErrQueryConfiguration = errors.New("billing usage query is misconfigured")
	ErrQueryScope         = errors.New("billing usage query scope is invalid")
	ErrQueryDependency    = errors.New("billing usage query dependency is unavailable")
	ErrQueryInvariant     = errors.New("billing usage query invariant failed")
	ErrQueryStorage       = errors.New("billing usage query storage failed")
)

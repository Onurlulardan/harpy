package discount

import "errors"

var (
	ErrQueryConfiguration = errors.New("billing discount query is misconfigured")
	ErrQueryScope         = errors.New("billing discount query scope is invalid")
	ErrEligibilityMissing = errors.New("billing discount is not eligible")
	ErrQueryInvariant     = errors.New("billing discount query invariant failed")
	ErrQueryDependency    = errors.New("billing discount query authority is unavailable")
	ErrQueryStorage       = errors.New("billing discount query storage failed")
)

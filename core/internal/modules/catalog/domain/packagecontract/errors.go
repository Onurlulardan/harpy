package packagecontract

import "errors"

var (
	ErrManifestInvalid   = errors.New("catalog package manifest is invalid")
	ErrManifestTampered  = errors.New("catalog package manifest attestation failed")
	ErrDependencyCycle   = errors.New("catalog package dependency cycle detected")
	ErrDependencyMissing = errors.New("catalog package dependency is unavailable")
)

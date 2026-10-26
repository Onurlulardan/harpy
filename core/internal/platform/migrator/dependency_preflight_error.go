package migrator

import (
	"errors"
	"fmt"
)

var ErrDependencyPreflight = errors.New("migration dependency preflight failed")

func dependencyPreflightError(cause error) error {
	if cause == nil {
		return ErrDependencyPreflight
	}
	return fmt.Errorf("%w: %w", ErrDependencyPreflight, cause)
}

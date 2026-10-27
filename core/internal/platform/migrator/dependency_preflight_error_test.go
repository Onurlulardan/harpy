package migrator

import (
	"errors"
	"testing"
)

func TestDependencyPreflightErrorPreservesTypedClassification(t *testing.T) {
	cause := errors.New("provider detail")
	err := dependencyPreflightError(cause)
	if !errors.Is(err, ErrDependencyPreflight) || !errors.Is(err, cause) {
		t.Fatalf("preflight classification lost: %v", err)
	}
}

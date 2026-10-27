package recoveryset

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) RecoverySetListing(
	context.Context, string,
) ([]platformopsapi.RecoverySet, error) {
	backend.called = true
	return []platformopsapi.RecoverySet{{RecoverySetID: "set"}}, nil
}

func TestListForwardsRecoverySetScope(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewRecoverySetOperations(backend).ListRecoverySets(
		t.Context(), "environment")
	if err != nil || !backend.called || len(result) != 1 {
		t.Fatalf("ListRecoverySets() = %#v, %v", result, err)
	}
}

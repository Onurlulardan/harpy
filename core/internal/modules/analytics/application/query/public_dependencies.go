package query

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *PublicApplication) resolveDependencies(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	queryID, versionID uuid.UUID,
) ([]DependencySnapshot, error) {
	pinned, err := store.Dependencies(ctx, queryID, versionID)
	if err != nil || len(pinned) == 0 || len(pinned) > 32 {
		return nil, ErrOwnerNotV2
	}
	requests := make([]DependencyRequest, len(pinned))
	for index, value := range pinned {
		requests[index] = value.DependencyRequest
	}
	current, err := application.dependencies.Dependencies(ctx, scope, requests)
	if err != nil || len(current) != len(pinned) {
		return nil, ErrOwnerNotV2
	}
	byIdentity := make(map[DependencyRequest]DependencySnapshot, len(current))
	for _, value := range current {
		if _, duplicate := byIdentity[value.DependencyRequest]; duplicate {
			return nil, ErrOwnerNotV2
		}
		byIdentity[value.DependencyRequest] = value
	}
	verified := make([]DependencySnapshot, len(pinned))
	for index, expected := range pinned {
		observed, found := byIdentity[expected.DependencyRequest]
		if !found || !samePinnedDependency(expected, observed) {
			return nil, ErrOwnerNotV2
		}
		verified[index] = observed
	}
	return verified, nil
}

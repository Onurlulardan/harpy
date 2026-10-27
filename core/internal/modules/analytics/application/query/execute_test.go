package query

import (
	"context"
	"testing"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type queryFixtureCache struct {
	read  func(CacheRequest) ([]byte, bool, error)
	wrote bool
}

func (cache *queryFixtureCache) Read(_ context.Context, request CacheRequest) ([]byte, bool, error) {
	return cache.read(request)
}
func (cache *queryFixtureCache) Write(context.Context, CacheRequest, []byte, time.Time) error {
	cache.wrote = true
	return nil
}
func (*queryFixtureCache) Invalidate(context.Context, uuid.UUID) error { return nil }

func TestExecuteBoundRejectsStaleRevisionCacheRace(t *testing.T) {
	freshBinding := ExecutionBinding{ReleaseGeneration: 3, AuthorizationSHA256: [32]byte{1},
		PredicateSHA256: [32]byte{2}, RevisionsSHA256: [32]byte{3}}
	cache := &queryFixtureCache{read: func(request CacheRequest) ([]byte, bool, error) {
		if request.Binding.RevisionsSHA256 == [32]byte{9} {
			return []byte(`{"stale":true}`), true, nil
		}
		return nil, false, nil
	}}
	store := &queryFixtureStore{getVersion: func(context.Context, uuid.UUID, uuid.UUID) (Snapshot, error) {
		return queryFixtureSnapshot(analyticsquery.VersionPublished,
			analyticspublication.CacheRevisionBound), nil
	}, dependencies: func(context.Context, uuid.UUID, uuid.UUID) ([]DependencySnapshot, error) {
		return []DependencySnapshot{executionDependency()}, nil
	}}
	application := queryFixtureApplication()
	application.contracts = queryFixtureContracts{binding: freshBinding,
		result: ExecutionResult{Body: []byte(`{"fresh":true}`)}}
	application.cache = cache
	application.dependencies = queryFixtureDependencies{values: []DependencySnapshot{executionDependency()}}
	result, err := application.executeBound(context.Background(), platformtx.TenantScope{}, store,
		queryFixtureIdentity(), ExecuteRequest{QueryID: queryFixtureID,
			VersionID: queryFixtureVersion}, queryFixtureDocument(`{}`))
	if err != nil || string(result.Body) != `{"fresh":true}` || !cache.wrote {
		t.Fatalf("execute result = %s, wrote = %v, err = %v", result.Body, cache.wrote, err)
	}
}

func executionDependency() DependencySnapshot {
	return DependencySnapshot{DependencyRequest: DependencyRequest{OwnerDomain: "forms",
		ResourceKind: "form", ResourceID: queryFixtureID, VersionID: queryFixtureVersion},
		CatalogVersion: 1, QueryContractVersion: 1, PolicyVersion: 1,
		OwnerGeneration: 1, WriterV2: true,
		DataRevisionCapability: "forms.revision.v1", DependencySHA256: [32]byte{7}}
}

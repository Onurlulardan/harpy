package query

import (
	"context"
	"errors"
	"testing"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type queryFixtureEffects struct {
	effect PublishEffect
	err    error
}

func (effects *queryFixtureEffects) QueryPublished(
	_ context.Context, _ platformtx.TenantScope, effect PublishEffect,
) error {
	effects.effect = effect
	return effects.err
}

func TestPublishBoundFailsTransactionWhenAsyncEffectFails(t *testing.T) {
	requestDependency := DependencyRequest{OwnerDomain: "forms", ResourceKind: "form",
		ResourceID: queryFixtureID, VersionID: queryFixtureVersion}
	resolved := DependencySnapshot{DependencyRequest: requestDependency,
		CatalogVersion: 1, QueryContractVersion: 1, PolicyVersion: 1,
		OwnerGeneration: 2, WriterV2: true,
		DataRevisionCapability: "forms.revision.v1", DependencySHA256: [32]byte{2}}
	store := &queryFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return queryFixtureSnapshot(analyticsquery.VersionDraft, analyticspublication.CacheNone), nil
	}, publish: func(_ context.Context, mutation PublishMutation) (Snapshot, bool, error) {
		result := queryFixtureSnapshot(analyticsquery.VersionPublished,
			analyticspublication.CacheNone)
		result.ManifestSHA256 = mutation.ManifestSHA256
		result.RootRevision, result.VersionNumber, result.VersionRevision = 2, 1, 2
		return result, false, nil
	}}
	effects := &queryFixtureEffects{err: errors.New("outbox unavailable")}
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureNext)
	application.dependencies = queryFixtureDependencies{values: []DependencySnapshot{resolved}}
	application.contracts = queryFixtureContracts{}
	application.effects = effects
	request := PublishRequest{QueryID: queryFixtureID, DraftVersionID: queryFixtureVersion,
		ExpectedRootRevision: 1, ExpectedDraftRevision: 1,
		Dependencies: []DependencyRequest{requestDependency}, IdempotencyKey: "query-effect-1"}

	_, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), OperationPublish, request)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("publish error = %v", err)
	}
	if effects.effect.QueryVersionID != queryFixtureVersion ||
		effects.effect.RootRevision != 2 || effects.effect.VersionNumber != 1 {
		t.Fatalf("effect = %+v", effects.effect)
	}
}

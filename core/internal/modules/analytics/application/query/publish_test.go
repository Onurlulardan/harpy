package query

import (
	"context"
	"errors"
	"testing"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type queryFixtureDependencies struct {
	values []DependencySnapshot
	err    error
}

func (resolver queryFixtureDependencies) Dependencies(context.Context, platformtx.TenantScope,
	[]DependencyRequest,
) ([]DependencySnapshot, error) {
	return resolver.values, resolver.err
}

func TestPublishBoundAcceptsReplayAndPersistsRevisionCachePolicy(t *testing.T) {
	requestDependency := DependencyRequest{OwnerDomain: "iam", ResourceKind: "form",
		ResourceID: queryFixtureID, VersionID: queryFixtureVersion}
	resolved := DependencySnapshot{DependencyRequest: requestDependency,
		CatalogVersion: 1, QueryContractVersion: 1, PolicyVersion: 1,
		OwnerGeneration: 2, WriterV2: true, DataRevisionCapability: "forms.revision.v1",
		DependencySHA256: [32]byte{2}}
	var captured PublishMutation
	store := &queryFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return queryFixtureSnapshot(analyticsquery.VersionDraft, analyticspublication.CacheNone), nil
	}, publish: func(_ context.Context, mutation PublishMutation) (Snapshot, bool, error) {
		captured = mutation
		result := queryFixtureSnapshot(analyticsquery.VersionPublished,
			analyticspublication.CacheRevisionBound)
		result.ManifestSHA256, result.RootRevision, result.VersionRevision = mutation.ManifestSHA256, 2, 2
		return result, true, nil
	}}
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureNext)
	application.entitlements = &queryFixtureEntitlements{decision: queryFixtureMutationDecision(billingapi.MeteredResourcePublish, true)}
	application.dependencies = queryFixtureDependencies{values: []DependencySnapshot{resolved}}
	application.contracts = queryFixtureContracts{}
	request := PublishRequest{QueryID: queryFixtureID, DraftVersionID: queryFixtureVersion,
		ExpectedRootRevision: 1, ExpectedDraftRevision: 1,
		Dependencies: []DependencyRequest{requestDependency}, RevisionBoundCache: true,
		IdempotencyKey: "query-publish-1"}
	result, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), OperationPublish, request)
	if err != nil || result.VersionState != analyticsquery.VersionPublished ||
		captured.CachePolicy != analyticspublication.CacheRevisionBound ||
		captured.ManifestSHA256 == [32]byte{} {
		t.Fatalf("publish result = %+v, mutation = %+v, err = %v", result, captured, err)
	}
}

func TestPublishBoundFailsClosedWhenDependencyOwnerUnavailable(t *testing.T) {
	store := &queryFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return queryFixtureSnapshot(analyticsquery.VersionDraft, analyticspublication.CacheNone), nil
	}, publish: func(context.Context, PublishMutation) (Snapshot, bool, error) {
		t.Fatal("publish called after dependency failure")
		return Snapshot{}, false, nil
	}}
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureNext)
	application.dependencies = queryFixtureDependencies{err: errors.New("owner fence unavailable")}
	application.contracts = queryFixtureContracts{}
	_, err := application.publishBound(context.Background(), platformtx.TenantScope{}, store,
		queryFixtureIdentity(), OperationPublish, PublishRequest{QueryID: queryFixtureID,
			DraftVersionID: queryFixtureVersion, ExpectedRootRevision: 1,
			ExpectedDraftRevision: 1, IdempotencyKey: "query-publish-2",
			Dependencies: []DependencyRequest{{OwnerDomain: "iam", ResourceKind: "form",
				ResourceID: queryFixtureID, VersionID: queryFixtureVersion}}})
	if !errors.Is(err, ErrOwnerNotV2) {
		t.Fatalf("publish error = %v", err)
	}
}

func TestPublishBoundDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	requestDependency := DependencyRequest{OwnerDomain: "iam", ResourceKind: "form",
		ResourceID: queryFixtureID, VersionID: queryFixtureVersion}
	resolved := DependencySnapshot{DependencyRequest: requestDependency,
		CatalogVersion: 1, QueryContractVersion: 1, PolicyVersion: 1,
		OwnerGeneration: 2, WriterV2: true,
		DataRevisionCapability: "forms.revision.v1", DependencySHA256: [32]byte{2}}
	storeCalled := false
	store := &queryFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return queryFixtureSnapshot(analyticsquery.VersionDraft,
			analyticspublication.CacheNone), nil
	}, publish: func(context.Context, PublishMutation) (Snapshot, bool, error) {
		storeCalled = true
		return Snapshot{}, false, nil
	}}
	decision := queryFixtureMutationDecision(billingapi.MeteredResourcePublish, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureNext)
	application.entitlements = &queryFixtureEntitlements{decision: decision}
	application.dependencies = queryFixtureDependencies{values: []DependencySnapshot{resolved}}
	application.contracts = queryFixtureContracts{}
	_, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), OperationPublish, PublishRequest{
			QueryID: queryFixtureID, DraftVersionID: queryFixtureVersion,
			ExpectedRootRevision: 1, ExpectedDraftRevision: 1,
			Dependencies:   []DependencyRequest{requestDependency},
			IdempotencyKey: "query-publish-denied-1"})
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("publish error=%v owner_called=%v", err, storeCalled)
	}
}

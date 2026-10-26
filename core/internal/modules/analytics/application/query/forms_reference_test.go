package query

import (
	"context"
	"errors"
	"testing"
	"time"

	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type queryReferenceProbe struct {
	bound    []analyticsforms.Source
	targets  [][]analyticsforms.Target
	released []analyticsforms.ID
	err      error
}

func (probe *queryReferenceProbe) BindVersion(
	_ context.Context, _ platformtx.TenantScope, source analyticsforms.Source,
	targets []analyticsforms.Target, _ time.Time,
) error {
	probe.bound = append(probe.bound, source)
	probe.targets = append(probe.targets, append([]analyticsforms.Target(nil), targets...))
	return probe.err
}

func TestPublishRegistersOnlyFormsDependencies(t *testing.T) {
	formsDependency := DependencySnapshot{
		DependencyRequest: DependencyRequest{
			OwnerDomain: "forms", ResourceKind: "form",
			ResourceID: queryFixtureID, VersionID: queryFixtureVersion,
		},
	}
	iamDependency := DependencySnapshot{
		DependencyRequest: DependencyRequest{
			OwnerDomain: "iam", ResourceKind: "principal",
			ResourceID: queryFixtureNext, VersionID: queryFixtureOperation,
		},
	}
	targets, err := formsReferenceTargets(
		[]DependencySnapshot{formsDependency, iamDependency},
	)
	if err != nil || len(targets) != 1 ||
		targets[0].DefinitionID != analyticsforms.ID(queryFixtureID) {
		t.Fatalf("mixed owner target projection differs: %#v err=%v", targets, err)
	}
}

func (probe *queryReferenceProbe) DecommissionSource(
	_ context.Context, _ platformtx.TenantScope, sourceID analyticsforms.ID,
	_ *analyticsforms.ID, _ time.Time,
) error {
	probe.released = append(probe.released, sourceID)
	return probe.err
}

func TestPublishFailsClosedWhenFormsReferenceReceiptIsUnavailable(t *testing.T) {
	dependency := DependencySnapshot{
		DependencyRequest: DependencyRequest{
			OwnerDomain: "forms", ResourceKind: "form",
			ResourceID: queryFixtureID, VersionID: queryFixtureVersion,
		},
		CatalogVersion: 1, QueryContractVersion: 1, PolicyVersion: 1,
		OwnerGeneration: 1, WriterV2: true,
		DataRevisionCapability: "forms.revision.v1",
		DependencySHA256:       [32]byte{1},
	}
	store := &queryFixtureStore{
		get: func(context.Context, uuid.UUID) (Snapshot, error) {
			return queryFixtureSnapshot(
				analyticsquery.VersionDraft, analyticspublication.CacheNone,
			), nil
		},
		publish: func(_ context.Context, mutation PublishMutation) (Snapshot, bool, error) {
			result := queryFixtureSnapshot(
				analyticsquery.VersionPublished, analyticspublication.CacheNone,
			)
			result.ManifestSHA256 = mutation.ManifestSHA256
			return result, false, nil
		},
	}
	probe := &queryReferenceProbe{err: errors.New("forged owner receipt")}
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureNext)
	application.dependencies = queryFixtureDependencies{values: []DependencySnapshot{dependency}}
	application.contracts = queryFixtureContracts{}
	application.references = probe
	_, err := application.publishBound(
		context.Background(), platformtx.TenantScope{}, store, queryFixtureIdentity(),
		OperationPublish, PublishRequest{
			QueryID: queryFixtureID, DraftVersionID: queryFixtureVersion,
			ExpectedRootRevision: 1, ExpectedDraftRevision: 1,
			IdempotencyKey: "query-reference-failure",
			Dependencies:   []DependencyRequest{dependency.DependencyRequest},
		},
	)
	if err == nil || len(probe.bound) != 1 ||
		probe.bound[0].VersionID != analyticsforms.ID(queryFixtureVersion) {
		t.Fatalf("publish reference failure was not fail-closed: %v %#v", err, probe.bound)
	}
}

func TestDeleteReleasesExactQueryReferenceSource(t *testing.T) {
	store := &queryFixtureStore{
		delete: func(context.Context, DeleteMutation) (Snapshot, bool, error) {
			result := queryFixtureSnapshot(
				analyticsquery.VersionPublished, analyticspublication.CacheNone,
			)
			result.Lifecycle = analyticsquery.LifecycleRetired
			result.RootRevision = 2
			return result, false, nil
		},
	}
	probe := &queryReferenceProbe{}
	application := queryFixtureApplication(queryFixtureOperation)
	application.references = probe
	_, err := application.deleteBound(
		context.Background(), platformtx.TenantScope{}, store, queryFixtureIdentity(),
		DeleteRequest{QueryID: queryFixtureID, ExpectedRootRevision: 1,
			IdempotencyKey: "query-reference-release"},
	)
	if err != nil || len(probe.released) != 1 ||
		probe.released[0] != analyticsforms.ID(queryFixtureID) {
		t.Fatalf("delete release = %#v, err=%v", probe.released, err)
	}
}

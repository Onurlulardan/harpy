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

func TestRevalidateDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	dependency := DependencyRequest{
		OwnerDomain: "forms", ResourceKind: "form",
		ResourceID: queryFixtureID, VersionID: queryFixtureVersion,
	}
	resolved := DependencySnapshot{
		DependencyRequest: dependency,
		CatalogVersion:    1, QueryContractVersion: 1, PolicyVersion: 1,
		OwnerGeneration: 2, WriterV2: true,
		DataRevisionCapability: "forms.revision.v1", DependencySHA256: [32]byte{2},
	}
	storeCalled := false
	store := &queryFixtureStore{
		get: func(context.Context, uuid.UUID) (Snapshot, error) {
			return queryFixtureSnapshot(analyticsquery.VersionDraft,
				analyticspublication.CacheNone), nil
		},
		publish: func(context.Context, PublishMutation) (Snapshot, bool, error) {
			storeCalled = true
			return Snapshot{}, false, nil
		},
	}
	decision := queryFixtureMutationDecision(billingapi.MeteredResourceRevalidate, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := queryFixtureApplication(queryFixtureOperation, queryFixtureNext)
	application.entitlements = &queryFixtureEntitlements{decision: decision}
	application.dependencies = queryFixtureDependencies{values: []DependencySnapshot{resolved}}
	application.contracts = queryFixtureContracts{}
	_, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, queryFixtureIdentity(), OperationRevalidate, PublishRequest{
			QueryID: queryFixtureID, DraftVersionID: queryFixtureVersion,
			ExpectedRootRevision: 1, ExpectedDraftRevision: 1,
			Dependencies:   []DependencyRequest{dependency},
			IdempotencyKey: "query-revalidate-denied-1",
		})
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("revalidate error=%v owner_called=%v", err, storeCalled)
	}
}

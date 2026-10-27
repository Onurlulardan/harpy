package dashboard

import (
	"context"
	"errors"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
	"testing"
)

func TestPublishBoundAcceptsReplayAndExactVerifiedManifest(t *testing.T) {
	manifest := [32]byte{8}
	store := &dashboardFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return dashboardFixtureSnapshot(false), nil
	}, publish: func(_ context.Context, mutation PublishMutation) (Snapshot, bool, error) {
		result := dashboardFixtureSnapshot(true)
		result.ManifestSHA256, result.RootRevision, result.VersionRevision = mutation.ManifestSHA256, 2, 2
		return result, true, nil
	}}
	application := &Application{queries: dashboardFixtureQueries{manifest: manifest},
		entitlements: &dashboardFixtureEntitlements{decision: dashboardFixtureMutationDecision(billingapi.MeteredResourcePublish, true)},
		ids:          &dashboardFixtureIDs{values: []uuid.UUID{dashboardFixtureOperation, dashboardFixtureNext}},
		clock:        dashboardFixtureClock{}}
	result, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), PublishRequest{DashboardID: dashboardFixtureID,
			DraftVersionID: dashboardFixtureVersion, ExpectedRootRevision: 1,
			ExpectedDraftRevision: 1, IdempotencyKey: "dashboard-publish-1"})
	if err != nil || !result.Published || result.ManifestSHA256 != manifest {
		t.Fatalf("publish result = %+v, err = %v", result, err)
	}
}

func TestPublishBoundDoesNotWriteOwnerStateWhenBuilderIsDisabled(t *testing.T) {
	manifest := [32]byte{8}
	storeCalled := false
	store := &dashboardFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return dashboardFixtureSnapshot(false), nil
	}, publish: func(context.Context, PublishMutation) (Snapshot, bool, error) {
		storeCalled = true
		return Snapshot{}, false, nil
	}}
	decision := dashboardFixtureMutationDecision(billingapi.MeteredResourcePublish, false)
	decision.FeatureEnabled, decision.Allowed = false, false
	application := &Application{queries: dashboardFixtureQueries{manifest: manifest},
		entitlements: &dashboardFixtureEntitlements{decision: decision},
		ids: &dashboardFixtureIDs{values: []uuid.UUID{
			dashboardFixtureOperation, dashboardFixtureNext,
		}}, clock: dashboardFixtureClock{}}
	_, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), PublishRequest{DashboardID: dashboardFixtureID,
			DraftVersionID: dashboardFixtureVersion, ExpectedRootRevision: 1,
			ExpectedDraftRevision: 1, IdempotencyKey: "dashboard-publish-denied-1"})
	if !errors.Is(err, ErrEntitlement) || storeCalled {
		t.Fatalf("publish error=%v owner_called=%v", err, storeCalled)
	}
}

package dashboard

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type dashboardFixtureEffects struct {
	effect PublishEffect
	err    error
}

func (effects *dashboardFixtureEffects) DashboardPublished(
	_ context.Context, _ platformtx.TenantScope, effect PublishEffect,
) error {
	effects.effect = effect
	return effects.err
}

func TestPublishBoundFailsTransactionWhenAsyncEffectFails(t *testing.T) {
	manifest := [32]byte{8}
	store := &dashboardFixtureStore{get: func(context.Context, uuid.UUID) (Snapshot, error) {
		return dashboardFixtureSnapshot(false), nil
	}, publish: func(_ context.Context, mutation PublishMutation) (Snapshot, bool, error) {
		result := dashboardFixtureSnapshot(true)
		result.ManifestSHA256 = mutation.ManifestSHA256
		result.RootRevision, result.VersionNumber, result.VersionRevision = 2, 1, 2
		return result, false, nil
	}}
	effects := &dashboardFixtureEffects{err: errors.New("outbox unavailable")}
	application := &Application{queries: dashboardFixtureQueries{manifest: manifest},
		entitlements: &dashboardFixtureEntitlements{},
		ids: &dashboardFixtureIDs{values: []uuid.UUID{
			dashboardFixtureOperation, dashboardFixtureNext,
		}}, clock: dashboardFixtureClock{}, effects: effects}
	request := PublishRequest{DashboardID: dashboardFixtureID,
		DraftVersionID: dashboardFixtureVersion, ExpectedRootRevision: 1,
		ExpectedDraftRevision: 1, IdempotencyKey: "dashboard-effect-1"}

	_, err := application.publishBound(context.Background(), platformtx.TenantScope{},
		store, dashboardFixtureIdentity(), request)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("publish error = %v", err)
	}
	if effects.effect.DashboardVersionID != dashboardFixtureVersion ||
		effects.effect.RootRevision != 2 || effects.effect.VersionNumber != 1 {
		t.Fatalf("effect = %+v", effects.effect)
	}
}

package billing

import (
	"context"
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func TestResolveLegacySubscriptionIdentityUsesTenantSnapshotAlias(t *testing.T) {
	fixture := newApplicationFixture(t)
	legacyID := uuid.New().String()
	fixture.bindings.current.View.LegacyID = legacyID
	resolved, err := fixture.application.ResolveLegacyIdentity(
		context.Background(), billingapi.LegacyIdentityQuery{
			Access: fixture.access, Kind: billingapi.LegacyIdentitySubscription,
			LegacyID: legacyID,
		},
	)
	if err != nil || resolved.LegacyID != legacyID ||
		resolved.CanonicalID != fixture.bindings.current.View.ID {
		t.Fatalf("resolved=%+v error=%v", resolved, err)
	}
}

func TestResolveLegacySubscriptionIdentityCanonicalV7Fallback(t *testing.T) {
	fixture := newApplicationFixture(t)
	canonicalID := fixture.bindings.current.View.ID
	resolved, err := fixture.application.ResolveLegacyIdentity(
		context.Background(), billingapi.LegacyIdentityQuery{
			Access: fixture.access, Kind: billingapi.LegacyIdentitySubscription,
			LegacyID: canonicalID,
		},
	)
	if err != nil || resolved.LegacyID != canonicalID ||
		resolved.CanonicalID != canonicalID {
		t.Fatalf("resolved=%+v error=%v", resolved, err)
	}
	_, err = fixture.application.ResolveLegacyIdentity(
		context.Background(), billingapi.LegacyIdentityQuery{
			Access: fixture.access, Kind: billingapi.LegacyIdentitySubscription,
			LegacyID: uuid.New().String(),
		},
	)
	if !errors.Is(err, billingapi.ErrNotFound) {
		t.Fatalf("missing identity error=%v", err)
	}
}

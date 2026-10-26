package api

import (
	"context"
	"testing"
	"time"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
)

func TestAnalyticsLegacyPublicDatabaseMappingIsLossless(t *testing.T) {
	last := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	acked := last.Add(time.Hour)
	want := analyticsmodule.LegacyPublicMapping{
		Key: analyticsmodule.LegacyPublicKey{TenantIdentity: "tenant-a", Method: "GET",
			Path: "/api/openapi/orders", Slug: "orders"},
		TenantID:         [16]byte{1},
		LegacyEndpointID: [16]byte{2},
		PublicID:         [16]byte{3},
		ReleaseID:        [16]byte{4},
		State:            "active", HitCount: 7, LastHitAt: &last, ConsumersAckedAt: &acked,
		RemoveAfter: acked.Add(30 * 24 * time.Hour),
	}
	got := analyticsLegacyPublicMapping(platformAnalyticsLegacyPublicRoute(want))
	if got != want {
		t.Fatalf("Analytics legacy mapping round trip = %+v, want %+v", got, want)
	}
}

func TestAnalyticsLegacyPublicDatabaseFailsClosedWithoutPool(t *testing.T) {
	database := analyticsLegacyPublicDatabase{}
	if _, err := database.ResolveAnalyticsLegacyPublic(
		context.Background(), analyticsmodule.LegacyPublicKey{},
	); err == nil {
		t.Fatal("nil Analytics legacy public database was accepted")
	}
}

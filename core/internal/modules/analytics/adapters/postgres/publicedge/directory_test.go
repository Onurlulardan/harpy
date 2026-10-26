package publicedge

import (
	"testing"

	publicationdomain "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestDirectoryResolvesOpaqueRouteAndMapsNotFound(t *testing.T) {
	publicID := [16]byte{1}
	database := &fixtureDatabase{route: platformpostgres.AnalyticsPublicRoute{
		PublicID: publicID, TenantID: uuid.New(),
		ReleaseID:         uuid.MustParse("019b0000-0000-7000-8000-000000000001"),
		ResourceKind:      "query",
		ResourceVersionID: uuid.MustParse("019b0000-0000-7000-8000-000000000002"),
		Audience:          string(publicationdomain.AudiencePublic), State: "active", Generation: 2,
	}}
	directory, err := NewDirectory(database)
	if err != nil {
		t.Fatal(err)
	}
	route, err := directory.Route(t.Context(), publicID)
	if err != nil || route.PublicID != publicID || route.TenantID != database.route.TenantID {
		t.Fatalf("route = %#v, err = %v", route, err)
	}
	database.resolveErr = &pgconn.PgError{Code: "P6102"}
	if _, err = directory.Route(t.Context(), publicID); err == nil {
		t.Fatal("missing route was accepted")
	}
}

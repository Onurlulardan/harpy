package publicedge

import (
	"context"
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type fixtureDatabase struct {
	route      platformpostgres.AnalyticsPublicRoute
	resolveErr error
	admitErr   error
	releaseErr error
	releases   int
}

func (database *fixtureDatabase) ResolveAnalyticsPublicRoute(
	context.Context, [16]byte,
) (platformpostgres.AnalyticsPublicRoute, error) {
	return database.route, database.resolveErr
}

func (database *fixtureDatabase) AdmitAnalyticsPublic(
	context.Context, uuid.UUID, uuid.UUID, int64, string, time.Time,
) (uuid.UUID, error) {
	return uuid.MustParse("019b0000-0000-7000-8000-000000000099"), database.admitErr
}

func (database *fixtureDatabase) ReleaseAnalyticsPublic(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int64,
) error {
	database.releases++
	return database.releaseErr
}

type fixtureClock struct{ now time.Time }

func (clock fixtureClock) Now() time.Time { return clock.now }

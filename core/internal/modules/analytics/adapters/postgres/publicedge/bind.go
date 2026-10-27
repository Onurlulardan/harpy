package publicedge

import (
	"context"
	"errors"
	"time"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type Database interface {
	ResolveAnalyticsPublicRoute(
		context.Context, [16]byte,
	) (platformpostgres.AnalyticsPublicRoute, error)
	AdmitAnalyticsPublic(
		context.Context, uuid.UUID, uuid.UUID, int64, string, time.Time,
	) (uuid.UUID, error)
	ReleaseAnalyticsPublic(
		context.Context, uuid.UUID, uuid.UUID, uuid.UUID, int64,
	) error
}

type Clock interface{ Now() time.Time }

type database = Database
type clock = Clock

type sqlStateError interface{ SQLState() string }

func mapResolveError(err error) error {
	var databaseError sqlStateError
	if errors.As(err, &databaseError) && databaseError.SQLState() == "P6102" {
		return publicationapp.ErrNotFound
	}
	return publicationapp.ErrUnavailable
}

func mapAdmissionError(err error) error {
	var databaseError sqlStateError
	if errors.As(err, &databaseError) && databaseError.SQLState() == "P6105" {
		return publicationapp.ErrRateLimited
	}
	return publicationapp.ErrRateLimitUnavailable
}

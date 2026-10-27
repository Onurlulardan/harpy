package publicedge

import (
	"context"
	"sync"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
)

type lease struct {
	database                     database
	leaseID, tenantID, releaseID uuid.UUID
	generation                   int64
	once                         sync.Once
	err                          error
}

func (value *lease) Close(ctx context.Context) error {
	if value == nil || value.database == nil || ctx == nil || ctx.Err() != nil {
		return publicationapp.ErrRateLimitUnavailable
	}
	value.once.Do(func() {
		value.err = value.database.ReleaseAnalyticsPublic(ctx, value.leaseID,
			value.tenantID, value.releaseID, value.generation)
		if value.err != nil {
			value.err = publicationapp.ErrRateLimitUnavailable
		}
	})
	return value.err
}

var _ publicationapp.PublicLease = (*lease)(nil)

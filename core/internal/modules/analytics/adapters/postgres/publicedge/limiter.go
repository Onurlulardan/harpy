package publicedge

import (
	"context"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

type Limiter struct {
	database database
	clock    clock
}

func NewLimiter(database database, clock clock) (*Limiter, error) {
	if nilValue(database) || nilValue(clock) {
		return nil, publicationapp.ErrInvalid
	}
	return &Limiter{database: database, clock: clock}, nil
}

func (limiter *Limiter) Admit(
	ctx context.Context, admission publicationapp.PublicAdmission,
) (publicationapp.PublicLease, error) {
	if limiter == nil || limiter.database == nil || limiter.clock == nil ||
		ctx == nil || ctx.Err() != nil {
		return nil, publicationapp.ErrRateLimitUnavailable
	}
	leaseID, err := limiter.database.AdmitAnalyticsPublic(ctx, admission.TenantID,
		admission.ReleaseID, int64(admission.Generation), admission.TrustedClientIP,
		limiter.clock.Now())
	if err != nil {
		return nil, mapAdmissionError(err)
	}
	return &lease{database: limiter.database, leaseID: leaseID,
		tenantID: admission.TenantID, releaseID: admission.ReleaseID,
		generation: int64(admission.Generation)}, nil
}

var _ publicationapp.PublicLimiter = (*Limiter)(nil)

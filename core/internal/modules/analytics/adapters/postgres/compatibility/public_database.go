package compatibility

import (
	"context"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

type PublicDatabase interface {
	ResolveAnalyticsLegacyPublic(
		context.Context, compatibilityapp.LegacyKey,
	) (compatibilityapp.LegacyMapping, error)
	RecordAnalyticsLegacyPublicHit(
		context.Context, compatibilityapp.LegacyMapping, time.Time,
	) (compatibilityapp.LegacyMapping, error)
}

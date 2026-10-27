package compatibility

import (
	"context"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func (store *Store) RecordHit(
	ctx context.Context, mapping compatibilityapp.LegacyMapping, at time.Time,
) (compatibilityapp.LegacyMapping, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		if store == nil || store.public == nil || ctx == nil || ctx.Err() != nil {
			return compatibilityapp.LegacyMapping{}, compatibilityapp.ErrUnavailable
		}
		result, err := store.public.RecordAnalyticsLegacyPublicHit(
			ctx, mapping, at,
		)
		if err != nil {
			return compatibilityapp.LegacyMapping{}, mapError(err)
		}
		return result, nil
	}
	arguments := append(mappingArgs(mapping), at)
	return scanMapping(store.database.QueryRow(ctx, recordHitSQL, arguments...))
}

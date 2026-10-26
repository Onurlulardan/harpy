package compatibility

import (
	"context"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func (store *Store) Resolve(
	ctx context.Context, key compatibilityapp.LegacyKey,
) (compatibilityapp.LegacyMapping, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		if store == nil || store.public == nil || ctx == nil || ctx.Err() != nil {
			return compatibilityapp.LegacyMapping{}, compatibilityapp.ErrUnavailable
		}
		result, err := store.public.ResolveAnalyticsLegacyPublic(ctx, key)
		if err != nil {
			return compatibilityapp.LegacyMapping{}, mapError(err)
		}
		return result, nil
	}
	return scanMapping(store.database.QueryRow(ctx, resolveSQL,
		key.TenantIdentity, key.Method, key.Path, key.Slug))
}

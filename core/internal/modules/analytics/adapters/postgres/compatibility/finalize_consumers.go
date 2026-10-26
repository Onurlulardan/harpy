package compatibility

import (
	"context"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func (store *Store) FinalizeConsumers(
	ctx context.Context, mapping compatibilityapp.LegacyMapping,
	inventorySHA256 [32]byte, at time.Time,
) (compatibilityapp.LegacyMapping, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return compatibilityapp.LegacyMapping{}, compatibilityapp.ErrUnavailable
	}
	arguments := append(mappingArgs(mapping), inventorySHA256[:], at)
	return scanMapping(store.database.QueryRow(ctx, acknowledgeConsumersSQL, arguments...))
}

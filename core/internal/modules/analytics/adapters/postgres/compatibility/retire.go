package compatibility

import (
	"context"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func (store *Store) Retire(
	ctx context.Context, mapping compatibilityapp.LegacyMapping,
	evidence compatibilityapp.RetirementEvidence, at time.Time,
) (compatibilityapp.LegacyMapping, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return compatibilityapp.LegacyMapping{}, compatibilityapp.ErrUnavailable
	}
	arguments := append(mappingArgs(mapping), evidence.SHA256[:], evidence.ObservedAt, at)
	return scanMapping(store.database.QueryRow(ctx, retireSQL, arguments...))
}

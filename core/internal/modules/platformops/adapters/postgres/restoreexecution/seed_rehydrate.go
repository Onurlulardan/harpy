package restoreexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) SeedRehydrateControl(
	ctx context.Context,
	record ports.RehydrateControlSeed,
) (bool, error) {
	seeded := false
	err := store.withBound(ctx, func(database boundDatabase) error {
		return database.QueryRow(
			ctx, seedRehydrateSQL,
			record.TrustSHA256, record.TrustBytes,
			record.PlanSHA256, record.PlanBytes,
			record.ManifestSHA256, record.ManifestBytes,
		).Scan(&seeded)
	})
	if err != nil {
		return false, fmt.Errorf("seed rehydrate control: %w", err)
	}
	return seeded, nil
}

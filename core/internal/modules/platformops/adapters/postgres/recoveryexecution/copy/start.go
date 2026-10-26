package copy

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) Start(
	ctx context.Context,
	identity ports.RecoveryExecutionIdentity,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, func(database boundDatabase) error {
		return database.QueryRow(
			ctx,
			startSQL,
			identity.RecoverySetID,
			identity.AttemptID,
			identity.LeaseToken,
		).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("start recovery copy: %w", err)
	}
	return changed, nil
}

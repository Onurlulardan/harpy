package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RequestRecoverySet(
	ctx context.Context,
	record ports.RecoverySetRequestRecord,
) (int64, error) {
	var sequence int64
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		err := database.QueryRow(ctx, requestRecoverySetSQL,
			record.RecoverySetID, record.OperationID, record.EnvironmentID,
			record.CreatedAt).Scan(&sequence)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return 0, fmt.Errorf("request recovery set: %w", err)
	}
	return sequence, nil
}

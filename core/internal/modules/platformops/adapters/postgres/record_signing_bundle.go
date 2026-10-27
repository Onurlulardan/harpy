package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	"github.com/google/uuid"
)

func (store *Store) RecordSigningBundle(
	ctx context.Context,
	record ports.SigningBundleRecord,
) error {
	if !validSigningBundle(record) {
		return ports.ErrConflict
	}
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		_, err := database.Execute(ctx, recordSigningBundleSQL,
			record.EnvironmentID, record.Generation, record.SHA256,
			record.BundleBytes, record.CreatedAt)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("record recovery signing bundle: %w", err)
	}
	return nil
}

func validSigningBundle(record ports.SigningBundleRecord) bool {
	_, err := uuid.Parse(record.EnvironmentID)
	return err == nil && record.Generation > 0 && len(record.SHA256) == 64 &&
		len(record.BundleBytes) > 0 && len(record.BundleBytes) <= 1024*1024 &&
		!record.CreatedAt.IsZero()
}

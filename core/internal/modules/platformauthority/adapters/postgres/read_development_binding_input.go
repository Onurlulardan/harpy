package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	"github.com/google/uuid"
)

func (store *Store) ReadDevelopmentBindingInput(
	ctx context.Context,
	environmentID string,
) (ports.DevelopmentBindingInput, error) {
	if _, err := uuid.Parse(environmentID); err != nil {
		return ports.DevelopmentBindingInput{}, fmt.Errorf(
			"development authority environment identity is invalid",
		)
	}
	var result ports.DevelopmentBindingInput
	err := store.withBound(ctx, transactionSnapshotQuery, func(bound boundStore) error {
		return bound.database.QueryRow(
			ctx, developmentBindingInputSQL, environmentID,
		).Scan(&result.OperationID, &result.RegistryBindingID,
			&result.FinalizationReceiptSHA256,
			&result.RecoveryTrustBundleSHA256)
	})
	if err != nil {
		return ports.DevelopmentBindingInput{}, fmt.Errorf(
			"read development authority binding input: %w", err,
		)
	}
	return result, nil
}

package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) ApplySignerRotation(
	ctx context.Context,
	value ports.SignerRotationRecord,
) error {
	if !validSignerRotation(value) {
		return ports.ErrConflict
	}
	now := store.clock.Now()
	if !validDatabaseInstant(now) {
		return ports.ErrConflict
	}
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, activateSignerSQL,
			value.OperationID, value.BindingID, value.ExpectedSignerGeneration,
			value.ExpectedSignerKID, value.NewSignerGeneration, value.NewSignerKID,
			value.NewSignerPublicKeySHA256,
			optionalText(value.NewRecoveryTrustBundleSHA256), now)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("activate platform authority signer: %w", err)
	}
	return nil
}

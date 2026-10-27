package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) FenceSigner(
	ctx context.Context,
	value ports.SignerFenceRecord,
) error {
	if !validSignerFence(value) {
		return ports.ErrConflict
	}
	now := store.clock.Now()
	if !validDatabaseInstant(now) {
		return ports.ErrConflict
	}
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		_, err := bound.database.Execute(ctx, fenceSignerSQL, value.BindingID,
			value.ExpectedSignerGeneration, value.ExpectedSignerKID,
			value.EvidenceSHA256, now)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("fence platform authority signer: %w", err)
	}
	return nil
}

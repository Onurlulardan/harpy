package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) LoadHead(
	ctx context.Context,
	key ports.HeadKey,
) (ports.ExternalHead, error) {
	result := ports.ExternalHead{Key: key}
	err := store.withBound(ctx, transactionSnapshotQuery, func(bound boundStore) error {
		return store.mapDatabaseError(bound.database.QueryRow(
			ctx, getHeadSQL, key.RegistryBindingID, string(key.ScopeKind), key.ScopeID,
		).Scan(
			&result.Key.RegistryBindingID, &result.Key.ScopeKind, &result.Key.ScopeID,
			&result.Generation, &result.StateSHA256, &result.AckSHA256,
			&result.RegistrySequence,
		))
	})
	if err != nil {
		return ports.ExternalHead{}, fmt.Errorf("get platform authority head: %w", err)
	}
	return result, nil
}

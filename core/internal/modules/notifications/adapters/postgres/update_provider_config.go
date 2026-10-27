package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) UpdateProviderConfig(ctx context.Context,
	value ports.ProviderConfig, expected int64,
) error {
	if !store.usable(ctx, true) || !validProviderConfig(value) || expected < 1 ||
		value.Revision != expected+1 {
		return ErrInvalid
	}
	metadata, err := encodeMetadata(value.Metadata)
	if err != nil {
		return ErrInvalid
	}
	defer clear(metadata)
	result, err := store.execute(ctx, updateProviderConfigSQL, value.ID.String(),
		value.VersionID.String(), value.SecretID.String(), string(value.Kind), value.Name,
		value.Enabled, expected, value.Revision, metadata, value.MetadataDigest[:])
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

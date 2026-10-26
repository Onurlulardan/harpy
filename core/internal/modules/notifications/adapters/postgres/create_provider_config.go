package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) CreateProviderConfig(ctx context.Context,
	value ports.ProviderConfig,
) error {
	if !store.usable(ctx, true) || !validProviderConfig(value) || value.Revision != 1 {
		return ErrInvalid
	}
	metadata, err := encodeMetadata(value.Metadata)
	if err != nil {
		return ErrInvalid
	}
	defer clear(metadata)
	_, err = store.execute(ctx, insertProviderConfigSQL, value.ID.String(),
		value.VersionID.String(), value.SecretID.String(), string(value.Kind), value.Name,
		value.Enabled, value.Revision, metadata, value.MetadataDigest[:])
	return err
}

func validProviderConfig(value ports.ProviderConfig) bool {
	return value.ID.Valid() && value.VersionID.Valid() && value.SecretID.Valid() &&
		(value.Kind == ports.ProviderSMTP || value.Kind == ports.ProviderTelegram) &&
		value.Name != "" && value.Revision > 0 && value.MetadataDigest != [32]byte{}
}

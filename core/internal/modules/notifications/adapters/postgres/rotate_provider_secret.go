package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) RotateProviderSecret(ctx context.Context,
	config ports.ProviderConfig, secret ports.EncryptedSecret, expected int64,
) error {
	if !store.usable(ctx, true) || !validProviderConfig(config) ||
		!validEncryptedSecret(secret) || config.SecretID != secret.ID || expected < 1 ||
		config.Revision != expected+1 {
		return ErrInvalid
	}
	metadata, err := encodeMetadata(config.Metadata)
	if err != nil {
		return ErrInvalid
	}
	defer clear(metadata)
	result, err := store.execute(ctx, rotateProviderSecretSQL, config.ID.String(),
		config.VersionID.String(), secret.ID.String(), string(config.Kind), config.Name,
		config.Enabled, expected, config.Revision, secret.Ciphertext, secret.Nonce,
		secret.WrappedDEK, secret.KEKRevision, secret.Digest[:], metadata,
		config.MetadataDigest[:])
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

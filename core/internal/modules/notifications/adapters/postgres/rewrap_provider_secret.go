package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) RewrapProviderSecret(ctx context.Context,
	secret ports.EncryptedSecret, expected int64,
) error {
	if !store.usable(ctx, true) || !validEncryptedSecret(secret) || expected < 1 ||
		secret.KEKRevision == expected {
		return ErrInvalid
	}
	result, err := store.execute(ctx, rewrapProviderSecretSQL, secret.ID.String(),
		secret.WrappedDEK, secret.KEKRevision, expected, secret.Ciphertext,
		secret.Nonce, secret.Digest[:])
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

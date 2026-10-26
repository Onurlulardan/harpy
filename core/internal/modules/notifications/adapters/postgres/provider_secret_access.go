package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) LoadProviderSecret(ctx context.Context,
	id domain.ID,
) (ports.EncryptedSecret, error) {
	if !store.usable(ctx, false) || !id.Valid() {
		return ports.EncryptedSecret{}, ErrInvalid
	}
	var value ports.EncryptedSecret
	value, err := scanProviderSecret(store.queryRow(ctx, getProviderSecretSQL, id.String()))
	if store.notFound(err) {
		return ports.EncryptedSecret{}, ports.ErrProviderNotFound
	}
	return value, err
}

func scanProviderSecret(row scanner) (ports.EncryptedSecret, error) {
	var value ports.EncryptedSecret
	var idText string
	var digest []byte
	err := row.Scan(&idText, &value.Ciphertext, &value.Nonce,
		&value.WrappedDEK, &value.KEKRevision, &digest)
	if err == nil {
		value.ID, err = decodeID(idText)
	}
	if err == nil {
		value.Digest, err = decodeDigest(digest)
	}
	return value, err
}

func (store *Store) StoreProviderSecret(ctx context.Context,
	value ports.EncryptedSecret,
) error {
	if !store.usable(ctx, true) || !validEncryptedSecret(value) {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertProviderSecretSQL, value.ID.String(),
		value.Ciphertext, value.Nonce, value.WrappedDEK, value.KEKRevision, value.Digest[:])
	return err
}

func validEncryptedSecret(value ports.EncryptedSecret) bool {
	return value.ID.Valid() && len(value.Ciphertext) >= 17 && len(value.Nonce) == 12 &&
		len(value.WrappedDEK) >= 28 && value.KEKRevision > 0 && value.Digest != [32]byte{}
}

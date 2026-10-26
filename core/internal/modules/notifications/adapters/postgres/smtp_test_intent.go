package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) FindSMTPTestIntent(ctx context.Context,
	key string,
) (ports.SMTPTestIntent, bool, error) {
	if !store.usable(ctx, false) || key == "" {
		return ports.SMTPTestIntent{}, false, ErrInvalid
	}
	value := ports.SMTPTestIntent{IdempotencyKey: key}
	var operationID, versionID, actorID string
	var digest []byte
	err := store.queryRow(ctx, findSMTPTestIntentSQL, key).Scan(&operationID,
		&versionID, &actorID, &digest, &value.ExpectedSettingsRevision,
		&value.RequestedAt)
	if store.notFound(err) {
		return ports.SMTPTestIntent{}, false, nil
	}
	if err == nil {
		value.OperationID, err = decodeID(operationID)
	}
	if err == nil {
		value.ProviderVersionID, err = decodeID(versionID)
	}
	if err == nil {
		value.ActorID, err = decodeID(actorID)
	}
	if err == nil {
		value.RequestDigest, err = decodeDigest(digest)
	}
	if err != nil || !validStoredSMTPTest(value) {
		return ports.SMTPTestIntent{}, false, ErrInvalid
	}
	return value, true, nil
}

func (store *Store) InsertSMTPTestIntent(ctx context.Context,
	value ports.SMTPTestIntent,
) error {
	if !store.usable(ctx, true) || !validNewSMTPTest(value) {
		return ErrInvalid
	}
	secret := value.EncryptedTarget
	result, err := store.execute(ctx, insertSMTPTestIntentSQL,
		value.OperationID.String(), value.IdempotencyKey, value.RequestDigest[:],
		value.ExpectedSettingsRevision, value.ProviderVersionID.String(),
		value.ActorID.String(), secret.Ciphertext, secret.Nonce, secret.WrappedDEK,
		secret.KEKRevision, secret.Digest[:], value.RequestedAt.UTC())
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

func validStoredSMTPTest(value ports.SMTPTestIntent) bool {
	return value.OperationID.IsV7() && value.ProviderVersionID.Valid() &&
		value.ActorID.Valid() && value.IdempotencyKey != "" &&
		len(value.IdempotencyKey) <= 128 && value.RequestDigest != [32]byte{} &&
		value.ExpectedSettingsRevision > 0 && !value.RequestedAt.IsZero()
}

func validNewSMTPTest(value ports.SMTPTestIntent) bool {
	return validStoredSMTPTest(value) && value.EncryptedTarget.ID == value.OperationID &&
		validEncryptedSecret(value.EncryptedTarget)
}

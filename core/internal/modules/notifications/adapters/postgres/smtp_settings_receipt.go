package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) FindSMTPSettingsReceipt(ctx context.Context,
	key string,
) (ports.SMTPSettingsReceipt, bool, error) {
	if !store.usable(ctx, false) || key == "" {
		return ports.SMTPSettingsReceipt{}, false, ErrInvalid
	}
	receipt := ports.SMTPSettingsReceipt{IdempotencyKey: key}
	var digest []byte
	var providerID string
	err := store.queryRow(ctx, findSMTPSettingsReceiptSQL, key).Scan(&digest,
		&receipt.ExpectedSettingsRevision, &receipt.ResultSettingsRevision,
		&providerID, &receipt.ProviderRevision)
	if store.notFound(err) {
		return ports.SMTPSettingsReceipt{}, false, nil
	}
	if err == nil {
		receipt.RequestDigest, err = decodeDigest(digest)
	}
	if err == nil {
		receipt.ProviderConfigID, err = decodeID(providerID)
	}
	if err != nil || !validSMTPReceipt(receipt) {
		return ports.SMTPSettingsReceipt{}, false, ErrInvalid
	}
	return receipt, true, nil
}

func (store *Store) InsertSMTPSettingsReceipt(ctx context.Context,
	receipt ports.SMTPSettingsReceipt,
) error {
	if !store.usable(ctx, true) || !validSMTPReceipt(receipt) {
		return ErrInvalid
	}
	result, err := store.execute(ctx, insertSMTPSettingsReceiptSQL,
		receipt.IdempotencyKey, receipt.RequestDigest[:],
		receipt.ExpectedSettingsRevision, receipt.ResultSettingsRevision,
		receipt.ProviderConfigID.String(), receipt.ProviderRevision)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrConflict
	}
	return nil
}

func validSMTPReceipt(value ports.SMTPSettingsReceipt) bool {
	return value.IdempotencyKey != "" && len(value.IdempotencyKey) <= 128 &&
		value.RequestDigest != [32]byte{} && value.ExpectedSettingsRevision > 0 &&
		value.ResultSettingsRevision == value.ExpectedSettingsRevision+1 &&
		value.ProviderConfigID.Valid() && value.ProviderRevision > 0
}

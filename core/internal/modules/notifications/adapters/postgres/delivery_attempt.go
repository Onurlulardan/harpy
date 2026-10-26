package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) InsertAttempt(ctx context.Context,
	value ports.DeliveryAttempt,
) error {
	if !store.usable(ctx, true) || !value.ID.Valid() || !value.DeliveryID.Valid() ||
		value.Number < 1 || value.Number > 8 || value.LeaseGeneration < 1 ||
		value.StartedAt.IsZero() || value.FinishedAt.Before(value.StartedAt) {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertDeliveryAttemptSQL, value.ID.String(),
		value.DeliveryID.String(), value.Number, value.LeaseGeneration,
		value.ProviderKind, value.ResponseClass, value.ProviderReferenceDigest[:],
		value.StatusCode, value.ErrorCode, value.StartedAt, value.FinishedAt)
	return err
}

func (store *Store) InsertReceipt(ctx context.Context,
	value ports.DeliveryReceipt,
) error {
	if !store.usable(ctx, true) || !value.ID.Valid() || !value.DeliveryID.Valid() ||
		value.ProviderReferenceDigest == [32]byte{} || value.Outcome == "" ||
		value.SignatureKeyID == "" || value.ReceivedAt.IsZero() {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertDeliveryReceiptSQL, value.ID.String(),
		value.DeliveryID.String(), value.ProviderReferenceDigest[:], value.Outcome,
		value.SignatureKeyID, value.ReceivedAt)
	return err
}

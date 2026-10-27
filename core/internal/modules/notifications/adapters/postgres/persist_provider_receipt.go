package postgres

import (
	"context"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) PersistProviderReceipt(ctx context.Context,
	value ports.AuthenticatedProviderReceipt,
) (bool, error) {
	if !store.usable(ctx, true) || !validProviderReceipt(value) {
		return false, ErrInvalid
	}
	var deliveryID string
	var replay bool
	err := store.queryRow(ctx, recordProviderReceiptSQL,
		value.ID.String(), value.DeliveryID.String(), value.ProviderReferenceDigest[:],
		string(value.Outcome), value.SignatureKeyID, value.ReceivedAt.UTC(),
		value.TargetDigest[:]).Scan(&deliveryID, &replay)
	if store.notFound(err) {
		return false, ErrConflict
	}
	if err != nil || deliveryID != value.DeliveryID.String() {
		return false, ErrConflict
	}
	return replay, nil
}

func validProviderReceipt(value ports.AuthenticatedProviderReceipt) bool {
	if !value.ID.Valid() || !value.DeliveryID.Valid() ||
		value.ProviderReferenceDigest == [32]byte{} || value.ReceivedAt.IsZero() ||
		value.SignatureKeyID == "" || len(value.SignatureKeyID) > 256 ||
		!utf8.ValidString(value.SignatureKeyID) {
		return false
	}
	switch value.Outcome {
	case ports.ReceiptDelivered:
		return true
	case ports.ReceiptBounced, ports.ReceiptComplained:
		return value.TargetDigest != [32]byte{}
	default:
		return false
	}
}

var _ ports.ProviderReceiptRecorder = (*Store)(nil)

package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) InsertEffect(ctx context.Context, value domain.EffectReceipt) error {
	if !store.usable(ctx, true) || value.Producer == "" || !value.RecipientID.Valid() ||
		!value.MessageID.Valid() || value.EffectKeyHash == [32]byte{} ||
		value.CanonicalDigest == [32]byte{} {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertEffectSQL, value.Producer, value.EffectKeyHash[:],
		value.RecipientID.String(), value.CanonicalDigest[:], value.MessageID.String(),
		value.LegacyReceiptReference, string(value.Outcome), value.CreatedAt, value.RetainUntil)
	return err
}

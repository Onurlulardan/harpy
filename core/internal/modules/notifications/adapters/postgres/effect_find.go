package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) FindEffect(ctx context.Context, producer string,
	hash [32]byte, recipient domain.ID,
) (domain.EffectReceipt, bool, error) {
	if !store.usable(ctx, false) || producer == "" || hash == [32]byte{} ||
		!recipient.Valid() {
		return domain.EffectReceipt{}, false, ErrInvalid
	}
	var value domain.EffectReceipt
	var recipientText, messageText string
	var keyHash, canonical []byte
	err := store.queryRow(ctx, findEffectSQL, producer, hash[:], recipient.String()).Scan(
		&value.Producer, &keyHash, &recipientText, &canonical, &messageText,
		&value.LegacyReceiptReference, &value.Outcome, &value.CreatedAt, &value.RetainUntil)
	if err != nil {
		if store.notFound(err) {
			return domain.EffectReceipt{}, false, nil
		}
		return domain.EffectReceipt{}, false, err
	}
	value.RecipientID, err = decodeID(recipientText)
	if err == nil {
		value.MessageID, err = decodeID(messageText)
	}
	if err == nil {
		value.EffectKeyHash, err = decodeDigest(keyHash)
	}
	if err == nil {
		value.CanonicalDigest, err = decodeDigest(canonical)
	}
	return value, err == nil, err
}

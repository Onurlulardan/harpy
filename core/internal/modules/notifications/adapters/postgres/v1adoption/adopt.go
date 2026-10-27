package v1adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
)

func (store *Store) Adopt(ctx context.Context,
	value app.MappedRecord,
) ([32]byte, bool, error) {
	if !store.usable(ctx) || !validMapped(value) {
		return [32]byte{}, false, app.ErrInvalid
	}
	var existing []byte
	err := store.queryRow(ctx, getAdoptedSQL, value.MessageID,
		value.RecipientID, value.EffectKeyHash[:]).Scan(&existing)
	if err == nil {
		if len(existing) != 32 {
			return [32]byte{}, false, app.ErrConflict
		}
		var digest [32]byte
		copy(digest[:], existing)
		return digest, false, nil
	}
	if !store.notFound(err) {
		return [32]byte{}, false, app.ErrUnavailable
	}
	if err = store.insertMapped(ctx, value); err != nil {
		return [32]byte{}, false, err
	}
	return value.CanonicalDigest, true, nil
}

func validMapped(value app.MappedRecord) bool {
	return value.MessageID != [16]byte{} && value.RecipientID != [16]byte{} &&
		value.NotificationType != "" && value.Category != "" && value.Subject != "" &&
		value.Body != "" && value.InboxRevision > 0 && !value.CreatedAt.IsZero() &&
		value.CanonicalDigest != [32]byte{} && value.EffectKeyHash != [32]byte{} &&
		(value.Channel == "" || value.Channel == "email" &&
			value.DeliveryID != [16]byte{} && value.AttemptID != [16]byte{})
}

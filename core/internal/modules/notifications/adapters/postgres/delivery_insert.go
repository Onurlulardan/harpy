package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) InsertDelivery(ctx context.Context, value domain.Delivery) error {
	if !store.usable(ctx, true) || !value.ID.Valid() || !value.MessageID.Valid() ||
		!value.Channel.Valid() || value.Channel == domain.ChannelInApp ||
		value.MaximumAttempts != 8 || value.Revision < 1 {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertDeliverySQL, value.ID.String(),
		value.MessageID.String(), nullableID(value.ProviderConfigID), string(value.Channel),
		string(value.State), value.Attempt, value.MaximumAttempts, value.LeaseOwner,
		value.LeaseGeneration, nullableTime(value.LeaseUntil), value.NotBefore,
		value.OutcomeCode, value.OutcomeUnknown, value.Mandatory, value.Revision)
	return err
}

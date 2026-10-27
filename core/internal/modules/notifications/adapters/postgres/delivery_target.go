package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

func (runtime *DeliveryRuntime) DeliveryTarget(ctx context.Context,
	tenantID, recipientID domain.ID, channel domain.Channel,
) ([]byte, error) {
	if runtime == nil || runtime.withStore == nil || ctx == nil || ctx.Err() != nil ||
		!tenantID.Valid() || !recipientID.Valid() || !channel.Valid() {
		return nil, ErrInvalid
	}
	tenant, err := uuid.Parse(tenantID.String())
	if err != nil {
		return nil, ErrInvalid
	}
	var target []byte
	err = runtime.withStore(ctx, tenant, func(scoped context.Context, store *Store) error {
		resolved, resolveErr := scanNotificationRecipient(store.queryRow(
			scoped, resolveNotificationRecipientSQL, recipientID.String()))
		if resolveErr != nil || resolved.principalID != recipientID {
			return ErrInvalid
		}
		target, resolveErr = resolved.target(channel)
		return resolveErr
	})
	return target, err
}

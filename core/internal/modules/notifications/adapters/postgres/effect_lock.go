package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func (store *Store) LockEffect(ctx context.Context, producer string,
	hash [32]byte, recipient domain.ID,
) error {
	if !store.usable(ctx, true) || producer == "" || hash == [32]byte{} ||
		!recipient.Valid() {
		return ErrInvalid
	}
	_, err := store.execute(ctx, lockEffectSQL, producer, hash[:], recipient.String())
	return err
}

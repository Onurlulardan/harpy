package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) InsertLegacyConnectionHit(ctx context.Context,
	value ports.LegacyConnectionHit,
) error {
	if !store.usable(ctx, true) || !value.TenantID.Valid() || !value.UserID.Valid() ||
		value.ConsumerManifestDigest == [32]byte{} {
		return ErrInvalid
	}
	_, err := store.execute(ctx, insertLegacyConnectionHitSQL,
		value.TenantID.String(), value.UserID.String(), value.ConsumerManifestDigest[:])
	return err
}

var _ ports.LegacyConnectionStore = (*Store)(nil)

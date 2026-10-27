package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
	"github.com/google/uuid"
)

func (store *Store) insertProviderLocator(ctx context.Context, tenantID uuid.UUID,
	value ports.ProviderLocator,
) error {
	if !store.usable(ctx, true) || tenantID == uuid.Nil || !value.DeliveryID.Valid() ||
		(value.ProviderKind != ports.ProviderSMTP && value.ProviderKind != ports.ProviderTelegram) ||
		value.ProviderAccountDigest == [32]byte{} || value.ExternalMessageDigest == [32]byte{} ||
		value.AcceptedAt.IsZero() {
		return ErrInvalid
	}
	var storedTenant, storedDelivery string
	err := store.queryRow(ctx, insertProviderLocatorSQL, string(value.ProviderKind),
		value.ProviderAccountDigest[:], value.ExternalMessageDigest[:], tenantID,
		value.DeliveryID.String(), value.AcceptedAt.UTC()).Scan(&storedTenant, &storedDelivery)
	if err != nil || storedTenant != tenantID.String() || storedDelivery != value.DeliveryID.String() {
		return ErrConflict
	}
	return nil
}

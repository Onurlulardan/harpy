package installationdelivery

import (
	"context"

	"github.com/google/uuid"
)

type ReadQuery struct{ OperationID uuid.UUID }

func Read(ctx context.Context, store Store, query ReadQuery) (Delivery, error) {
	if unusable(ctx, store) || query.OperationID.Version() != uuid.Version(7) {
		return Delivery{}, ErrInvalid
	}
	delivery, err := store.Read(ctx, query)
	if err != nil {
		return Delivery{}, err
	}
	if !delivery.Valid() || delivery.OperationID != query.OperationID {
		return Delivery{}, ErrUnavailable
	}
	return delivery, nil
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (store *Store) ClaimDelivery(
	ctx context.Context,
	leaseOwnerID string,
) (ports.DeliveryClaim, bool, error) {
	if strings.TrimSpace(leaseOwnerID) == "" {
		return ports.DeliveryClaim{}, false, ports.ErrConflict
	}
	var claim ports.DeliveryClaim
	err := store.withBound(ctx, transactionCommand, func(bound boundStore) error {
		err := bound.database.QueryRow(ctx, claimDeliverySQL, leaseOwnerID).Scan(
			&claim.OperationID, &claim.LeaseGeneration,
			&claim.Attempt, &claim.MaxAttempts,
		)
		return store.mapDatabaseError(err)
	})
	if errors.Is(err, ports.ErrNotFound) {
		return ports.DeliveryClaim{}, false, nil
	}
	if err != nil {
		return ports.DeliveryClaim{}, false,
			fmt.Errorf("claim platform authority delivery: %w", err)
	}
	claim.LeaseOwnerID = leaseOwnerID
	return claim, true, nil
}

var _ ports.DeliveryStore = (*Store)(nil)

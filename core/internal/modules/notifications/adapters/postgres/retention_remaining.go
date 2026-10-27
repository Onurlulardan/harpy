package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (store *RetentionStore) hasRemaining(ctx context.Context, tenantID uuid.UUID,
	now time.Time,
) (bool, error) {
	var tenantRemaining bool
	if err := store.database.QueryRow(
		ctx, hasRemainingTenantSQL, now,
	).Scan(&tenantRemaining); err != nil {
		return false, fmt.Errorf("inspect notification retention backlog: %w", err)
	}
	var platformRemaining bool
	if err := store.database.QueryRow(
		ctx, hasRemainingPlatformSQL, tenantID, now,
	).Scan(&platformRemaining); err != nil {
		return false, fmt.Errorf("inspect notification platform retention backlog: %w", err)
	}
	return tenantRemaining || platformRemaining, nil
}

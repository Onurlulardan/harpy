package retention

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (storage *store) authorityID(
	ctx context.Context, tenantID uuid.UUID,
) (uuid.UUID, error) {
	if storage == nil || storage.database.QueryRow == nil || ctx == nil ||
		ctx.Err() != nil || tenantID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("billing retention authority input is invalid")
	}
	var authorityID uuid.UUID
	if err := storage.database.QueryRow(
		ctx, resolveAuthoritySQL, tenantID,
	).Scan(&authorityID); err != nil || authorityID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("resolve billing retention authority: %w", err)
	}
	return authorityID, nil
}

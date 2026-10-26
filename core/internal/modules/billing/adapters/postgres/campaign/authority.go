package campaign

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (storage *store) authorityID(ctx context.Context) (uuid.UUID, error) {
	var identifier uuid.UUID
	if err := storage.queryRow(ctx, resolveAuthoritySQL).Scan(&identifier); err != nil {
		return uuid.Nil, fmt.Errorf("resolve billing global authority: %w", err)
	}
	if identifier == uuid.Nil {
		return uuid.Nil, fmt.Errorf("resolve billing global authority: empty identifier")
	}
	return identifier, nil
}

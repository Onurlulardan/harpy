package discount

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	"github.com/google/uuid"
)

func (storage *store) authorityID(ctx context.Context) (uuid.UUID, error) {
	if !storage.usable(ctx) {
		return uuid.Nil, discountapp.ErrConfiguration
	}
	var authorityID uuid.UUID
	if err := storage.queryRow(ctx, resolveAuthoritySQL, storage.tenantID).Scan(
		&authorityID,
	); err != nil {
		return uuid.Nil, fmt.Errorf("resolve billing discount authority: %w", err)
	}
	if err := storage.execute(ctx, assertAuthoritySQL, authorityID); err != nil {
		return uuid.Nil, fmt.Errorf("assert billing discount authority: %w", err)
	}
	return authorityID, nil
}

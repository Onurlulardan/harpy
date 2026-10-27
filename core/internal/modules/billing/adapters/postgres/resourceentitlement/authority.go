package resourceentitlement

import (
	"context"
	"fmt"

	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/google/uuid"
)

func (storage *store) authorityID(ctx context.Context) (uuid.UUID, error) {
	if !storage.usable(ctx) {
		return uuid.Nil, resourceapp.ErrConfiguration
	}
	var authorityID uuid.UUID
	if err := storage.queryRow(ctx, resolveAuthoritySQL, storage.tenantID).Scan(
		&authorityID,
	); err != nil {
		return uuid.Nil, fmt.Errorf("resolve metered resource authority: %w", err)
	}
	if err := storage.execute(ctx, assertAuthoritySQL, authorityID); err != nil {
		return uuid.Nil, fmt.Errorf("assert metered resource authority: %w", err)
	}
	return authorityID, nil
}

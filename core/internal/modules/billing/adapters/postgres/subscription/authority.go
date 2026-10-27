package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) authorityID(
	ctx context.Context,
	mode bindingMode,
) (uuid.UUID, error) {
	if !storage.usable(ctx, mode) {
		return uuid.Nil, subscriptionapp.ErrConfiguration
	}
	var authorityID uuid.UUID
	if err := storage.queryRow(ctx, resolveAuthoritySQL, storage.tenantID).Scan(
		&authorityID,
	); err != nil {
		return uuid.Nil, fmt.Errorf("resolve billing subscription authority: %w", err)
	}
	if err := storage.execute(ctx, assertAuthoritySQL, authorityID); err != nil {
		return uuid.Nil, fmt.Errorf("assert billing subscription authority: %w", err)
	}
	return authorityID, nil
}

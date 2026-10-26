package resourceentitlement

import (
	"context"
	"fmt"

	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func (storage *store) ResolveFeature(
	ctx context.Context,
	key entitlement.Key,
) (bool, error) {
	if !storage.usable(ctx) || !key.Valid() {
		return false, resourceapp.ErrRequest
	}
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return false, err
	}
	var enabled bool
	if err := storage.queryRow(ctx, resolveFeatureSQL, authorityID,
		storage.tenantID, string(key),
	).Scan(&enabled); err != nil {
		return false, fmt.Errorf("resolve metered resource feature: %w", err)
	}
	return enabled, nil
}

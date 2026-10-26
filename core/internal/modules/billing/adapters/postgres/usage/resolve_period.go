package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

func (storage *store) ResolvePeriod(
	ctx context.Context,
	key entitlement.Key,
) (uuid.UUID, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	var periodID uuid.UUID
	err = storage.queryRow(
		ctx, resolvePeriodSQL, authorityID, storage.tenantID, string(key),
	).Scan(&periodID)
	if storage.notFound(err) {
		return uuid.Nil, usageapp.ErrSnapshot
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve billing usage period: %w", err)
	}
	if periodID == uuid.Nil || periodID.Version() != uuid.Version(7) {
		return uuid.Nil, usageapp.ErrSnapshot
	}
	return periodID, nil
}

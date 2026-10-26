package usage

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

func (commands *Commands) ResolvePeriod(
	ctx context.Context,
	key entitlement.Key,
) (uuid.UUID, error) {
	if commands == nil || commands.store == nil {
		return uuid.Nil, ErrConfiguration
	}
	if ctx == nil || ctx.Err() != nil || !key.Valid() {
		return uuid.Nil, ErrRequest
	}
	periodID, err := commands.store.ResolvePeriod(ctx, key)
	if err != nil {
		return uuid.Nil, usageOperationError("resolve billing usage period", err)
	}
	if !validV7(periodID) {
		return uuid.Nil, ErrSnapshot
	}
	return periodID, nil
}

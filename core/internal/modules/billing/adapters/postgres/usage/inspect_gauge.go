package usage

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

func (storage *store) InspectGauge(
	ctx context.Context,
	key entitlement.Key,
	periodID uuid.UUID,
) (usageapp.GaugeSnapshot, error) {
	if !storage.usable(ctx) || !key.Valid() ||
		periodID == uuid.Nil || periodID.Version() != uuid.Version(7) {
		return usageapp.GaugeSnapshot{}, usageapp.ErrConfiguration
	}
	return storage.lockGauge(ctx, key, periodID)
}

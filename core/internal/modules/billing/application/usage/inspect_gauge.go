package usage

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func (commands *Commands) InspectGauge(
	ctx context.Context,
	key entitlement.Key,
) (GaugeSnapshot, error) {
	if commands == nil || !key.Valid() {
		return GaugeSnapshot{}, ErrRequest
	}
	inspector, ok := commands.store.(GaugeInspector)
	if !ok || inspector == nil {
		return GaugeSnapshot{}, ErrConfiguration
	}
	periodID, err := commands.ResolvePeriod(ctx, key)
	if err != nil {
		return GaugeSnapshot{}, err
	}
	snapshot, err := inspector.InspectGauge(ctx, key, periodID)
	if err != nil {
		return GaugeSnapshot{}, usageOperationError("inspect billing usage gauge", err)
	}
	if !validGaugeSnapshot(snapshot, AdjustGaugeRequest{
		EntitlementKey: key,
		PeriodID:       periodID,
		OwnerFactID:    periodID,
		Delta:          1,
	}) {
		return GaugeSnapshot{}, ErrSnapshot
	}
	return snapshot, nil
}

package resourceusage

import (
	"context"

	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

type usageBridge struct {
	commands *usageapp.Commands
}

func Bind(commands *usageapp.Commands) (resourceapp.Usage, error) {
	if commands == nil {
		return nil, resourceapp.ErrConfiguration
	}
	return usageBridge{commands: commands}, nil
}

func (usage usageBridge) InspectGauge(
	ctx context.Context,
	key entitlement.Key,
) (resourceapp.GaugeSnapshot, error) {
	snapshot, err := usage.commands.InspectGauge(ctx, key)
	if err != nil {
		return resourceapp.GaugeSnapshot{}, err
	}
	return resourceapp.GaugeSnapshot{
		Definition: snapshot.Definition,
		Limit:      snapshot.Limit,
		PeriodID:   snapshot.PeriodID,
		Current:    snapshot.Current,
	}, nil
}

func (usage usageBridge) AdjustGauge(
	ctx context.Context,
	request resourceapp.GaugeAdjustmentRequest,
) (resourceapp.GaugeAdjustment, error) {
	result, err := usage.commands.AdjustGauge(ctx, usageapp.AdjustGaugeRequest{
		EntitlementKey: request.EntitlementKey,
		PeriodID:       request.PeriodID,
		OwnerFactID:    request.OwnerFactID,
		Delta:          request.Delta,
	})
	if err != nil {
		return resourceapp.GaugeAdjustment{}, err
	}
	return resourceapp.GaugeAdjustment{
		DecisionID: result.DecisionID,
		Entry: resourceapp.GaugeEntry{
			ID:             result.Entry.ID,
			DecisionID:     result.Entry.DecisionID,
			EntitlementKey: result.Entry.EntitlementKey,
			PeriodID:       result.Entry.PeriodID,
			OwnerFactID:    result.Entry.OwnerFactID,
			Delta:          result.Entry.Delta,
			Before:         result.Entry.Before,
			After:          result.Entry.After,
		},
		Allowed: result.Allowed,
		Replay:  result.Replay,
		Before:  result.Before,
		After:   result.After,
		Limit:   result.Limit,
	}, nil
}

package manual

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func (adapter *Adapter) Reconcile(
	ctx context.Context,
	request effectsv1.Request,
) (providerapp.EffectResult, error) {
	if ctx == nil || ctx.Err() != nil {
		return providerapp.EffectResult{}, providerapp.ErrProviderDependency
	}
	return adapter.receipt(request, ReconciliationOperation, "MANUAL_STATE_CONFIRMED")
}

package billing

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ProviderEffectReconciliation struct {
	identity   effectexecution.Identity
	reconciler providerapp.Reconciler
}

func NewProviderEffectReconciliation(
	registry providerapp.Registry,
	providerName string,
	operation string,
) (*ProviderEffectReconciliation, error) {
	reconciler, err := registry.ReconciliationFor(providerName)
	if err != nil {
		return nil, err
	}
	identity, err := effectexecution.NewIdentity(providerName, operation)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	return &ProviderEffectReconciliation{
		identity: identity, reconciler: reconciler,
	}, nil
}

func (reconciliation *ProviderEffectReconciliation) Identity() effectexecution.Identity {
	if reconciliation == nil {
		return effectexecution.Identity{}
	}
	return reconciliation.identity
}

func (reconciliation *ProviderEffectReconciliation) Execute(
	ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	if reconciliation == nil || reconciliation.reconciler == nil ||
		!transaction.NetworkAllowed(ctx) {
		return effectexecution.Result{}
	}
	result, err := reconciliation.reconciler.Reconcile(ctx, request)
	if err != nil || !result.ValidFor(request) {
		return effectexecution.Result{}
	}
	return providerPlatformResult(result)
}

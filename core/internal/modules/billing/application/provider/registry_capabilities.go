package providerapp

import "sort"

func (registry Registry) CommandProviders() []string {
	return registry.providersWith(func(adapter Adapter) bool {
		_, ok := adapter.(EffectExecutor)
		return ok && adapter.Descriptor().Capabilities().Commands
	})
}

func (registry Registry) ReconciliationProviders() []string {
	return registry.providersWith(func(adapter Adapter) bool {
		_, ok := adapter.(Reconciler)
		return ok && adapter.Descriptor().Capabilities().Reconciliation
	})
}

func (registry Registry) WebhookProviders() []string {
	return registry.providersWith(func(adapter Adapter) bool {
		_, ok := adapter.(WebhookAdapter)
		return ok && adapter.Descriptor().Capabilities().Webhooks
	})
}

func (registry Registry) providersWith(accept func(Adapter) bool) []string {
	if registry.adapters == nil || accept == nil {
		return nil
	}
	providers := make([]string, 0, len(registry.adapters))
	for key, adapter := range registry.adapters {
		if !nilAdapter(adapter) && accept(adapter) {
			providers = append(providers, string(key))
		}
	}
	sort.Strings(providers)
	return providers
}

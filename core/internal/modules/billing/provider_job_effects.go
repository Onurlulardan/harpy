package billing

import (
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (runtime *ProviderJobRuntime) effects() (effectexecution.Registry, error) {
	providers := make([]effectexecution.Provider, 0,
		len(runtime.registry.CommandProviders())+
			len(runtime.registry.ReconciliationProviders()))
	for _, name := range runtime.registry.CommandProviders() {
		command, err := NewProviderEffectCommand(
			runtime.registry, name, providerapp.CommandOperation,
		)
		if err != nil {
			return effectexecution.Registry{},
				fmt.Errorf("build billing provider command effect: %w", err)
		}
		providers = append(providers, command)
	}
	for _, name := range runtime.registry.ReconciliationProviders() {
		reconcile, err := NewProviderEffectReconciliation(
			runtime.registry, name, providerapp.ReconciliationOperation,
		)
		if err != nil {
			return effectexecution.Registry{},
				fmt.Errorf("build billing provider reconcile effect: %w", err)
		}
		providers = append(providers, reconcile)
	}
	registry, err := effectexecution.NewRegistry(providers)
	if err != nil {
		return effectexecution.Registry{},
			fmt.Errorf("build billing provider effect registry: %w", err)
	}
	return registry, nil
}

package providerapp

import (
	"reflect"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
)

type Registry struct {
	adapters map[domainprovider.Key]Adapter
}

func NewRegistry(adapters ...Adapter) (Registry, error) {
	registered := make(map[domainprovider.Key]Adapter, len(adapters))
	for _, adapter := range adapters {
		if nilAdapter(adapter) || !capabilitiesMatch(adapter) {
			return Registry{}, ErrConfiguration
		}
		key := adapter.Descriptor().Key()
		if _, duplicate := registered[key]; duplicate {
			return Registry{}, ErrConfiguration
		}
		registered[key] = adapter
	}
	manual, exists := registered[domainprovider.Manual]
	if !exists || manual.Descriptor() != domainprovider.ManualDescriptor() {
		return Registry{}, ErrConfiguration
	}
	return Registry{adapters: registered}, nil
}

func (registry Registry) Webhook(key domainprovider.Key) (WebhookAdapter, error) {
	adapter, err := registry.resolve(key)
	if err != nil {
		return nil, err
	}
	webhook, ok := adapter.(WebhookAdapter)
	if !ok || !adapter.Descriptor().Capabilities().Webhooks {
		return nil, ErrCapability
	}
	return webhook, nil
}

func (registry Registry) Effect(key domainprovider.Key) (EffectExecutor, error) {
	adapter, err := registry.resolve(key)
	if err != nil {
		return nil, err
	}
	executor, ok := adapter.(EffectExecutor)
	if !ok || !adapter.Descriptor().Capabilities().Commands {
		return nil, ErrCapability
	}
	return executor, nil
}

func (registry Registry) EffectFor(name string) (EffectExecutor, error) {
	key, err := domainprovider.NewKey(name)
	if err != nil {
		return nil, ErrCapability
	}
	return registry.Effect(key)
}

func (registry Registry) ReconciliationAdapter(key domainprovider.Key) (Reconciler, error) {
	adapter, err := registry.resolve(key)
	if err != nil {
		return nil, err
	}
	reconciler, ok := adapter.(Reconciler)
	if !ok || !adapter.Descriptor().Capabilities().Reconciliation {
		return nil, ErrCapability
	}
	return reconciler, nil
}

func (registry Registry) ReconciliationFor(name string) (Reconciler, error) {
	key, err := domainprovider.NewKey(name)
	if err != nil {
		return nil, ErrCapability
	}
	return registry.ReconciliationAdapter(key)
}

func (registry Registry) resolve(key domainprovider.Key) (Adapter, error) {
	if registry.adapters == nil || !key.Valid() {
		return nil, ErrCapability
	}
	adapter, exists := registry.adapters[key]
	if !exists || nilAdapter(adapter) || adapter.Descriptor().Key() != key {
		return nil, ErrCapability
	}
	return adapter, nil
}

func capabilitiesMatch(adapter Adapter) bool {
	descriptor := adapter.Descriptor()
	capabilities := descriptor.Capabilities()
	_, webhooks := adapter.(WebhookAdapter)
	_, commands := adapter.(EffectExecutor)
	_, reconciliation := adapter.(Reconciler)
	return descriptor.Key().Valid() && capabilities.Webhooks == webhooks &&
		capabilities.Commands == commands && capabilities.Reconciliation == reconciliation
}

func nilAdapter(adapter Adapter) bool {
	if adapter == nil {
		return true
	}
	value := reflect.ValueOf(adapter)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

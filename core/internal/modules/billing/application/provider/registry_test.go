package providerapp

import (
	"testing"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
)

func TestRegistryRequiresExactCapabilityImplementations(t *testing.T) {
	if _, err := NewRegistry(); err != ErrConfiguration {
		t.Fatalf("missing manual error=%v", err)
	}
	_, external := externalFixture(t)
	mismatch := &capabilityMismatchAdapter{externalFixtureAdapter: external}
	if _, err := NewRegistry(&manualFixtureAdapter{}, mismatch); err != ErrConfiguration {
		t.Fatalf("capability mismatch error=%v", err)
	}
}

func TestRegistryResolvesEachDeclaredProviderCapability(t *testing.T) {
	registry, key, external := registryFixture(t)
	webhook, err := registry.Webhook(key)
	if err != nil || webhook != external {
		t.Fatalf("webhook=%T error=%v", webhook, err)
	}
	if _, err := registry.Effect(key); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.ReconciliationAdapter(key); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Webhook(domainprovider.Manual); err != ErrCapability {
		t.Fatalf("manual webhook error=%v", err)
	}
}

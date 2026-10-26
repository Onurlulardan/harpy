package billing

import (
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func TestManualProviderRuntimeHasFinalCommandAndReconciliationCapabilities(t *testing.T) {
	registry, err := NewManualProviderRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if got := registry.CommandProviders(); len(got) != 1 || got[0] != "manual" {
		t.Fatalf("manual command providers = %v", got)
	}
	if got := registry.ReconciliationProviders(); len(got) != 1 || got[0] != "manual" {
		t.Fatalf("manual reconciliation providers = %v", got)
	}
	runtime, err := NewProviderJobRuntime(registry)
	if err != nil {
		t.Fatal(err)
	}
	effects, err := runtime.effects()
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{
		providerapp.CommandOperation,
		providerapp.ReconciliationOperation,
	} {
		identity, err := effectexecution.NewIdentity("manual", operation)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := effects.Resolve(identity); err != nil {
			t.Fatalf("manual %s effect is missing: %v", operation, err)
		}
	}
}

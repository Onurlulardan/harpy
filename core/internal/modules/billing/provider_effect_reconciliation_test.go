package billing

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/provider/manual"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func TestProviderEffectReconciliationRunsThroughExternalEffectRegistry(t *testing.T) {
	reconciliation, err := NewProviderEffectReconciliation(
		providerRegistry(t), "manual", manual.ReconciliationOperation,
	)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := effectexecution.NewRegistry(
		[]effectexecution.Provider{reconciliation},
	)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := registry.Resolve(reconciliation.Identity())
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolved.Invoke(
		context.Background(), providerEffectRequest(t, manual.ReconciliationOperation),
	)
	if err != nil || result.IsUnknown() {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "MANUAL_STATE_CONFIRMED" {
		t.Fatalf("receipt=%+v", receipt)
	}
}

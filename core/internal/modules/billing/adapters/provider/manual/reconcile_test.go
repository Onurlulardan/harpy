package manual

import (
	"context"
	"testing"
)

func TestReconcileConfirmsManualCanonicalArtifact(t *testing.T) {
	adapter, _ := New(manualClock{manualNow})
	request := manualRequest(t, ReconciliationOperation)
	result, err := adapter.Reconcile(context.Background(), request)
	if err != nil || result.IsUnknown() {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "MANUAL_STATE_CONFIRMED" ||
		receipt.ProviderReference == nil ||
		*receipt.ProviderReference != "manual/"+manualEffectID {
		t.Fatalf("receipt=%+v", receipt)
	}
}

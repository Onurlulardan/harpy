package compatibility

import (
	"context"
	"testing"
)

func TestAcceptRequiresExactDispositionCounts(t *testing.T) {
	application, store := applicationFixture(t, mappingFixture())
	receipt := Acceptance{RunID: runID, TenantID: tenantID, CutoverGeneration: 7,
		SourceCount: 3, AdoptedCount: 2, RetiredCount: 1, RootSHA256: [32]byte{3}}
	result, err := application.Accept(context.Background(), authority, receipt)
	if err != nil || result.AcceptedAt != compatNow || store.receipt != result {
		t.Fatalf("result=%+v stored=%+v error=%v", result, store.receipt, err)
	}
	receipt.SourceCount++
	if _, err = application.Accept(context.Background(), authority, receipt); err != ErrInvalid {
		t.Fatalf("count error=%v", err)
	}
}

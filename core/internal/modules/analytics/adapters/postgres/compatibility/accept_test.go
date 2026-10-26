package compatibility

import (
	"context"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	"github.com/google/uuid"
)

func TestAcceptScansPersistedExactCountsAndDigest(t *testing.T) {
	receipt := compatibilityapp.Acceptance{RunID: testRun, TenantID: testTenant,
		CutoverGeneration: 8, SourceCount: 2, AdoptedCount: 1, RetiredCount: 1,
		RootSHA256: [32]byte{7}, AcceptedAt: testTime}
	store, _ := newStore(t, scanRow(func(targets []any) error {
		*targets[0].(*uuid.UUID) = receipt.RunID
		*targets[1].(*uuid.UUID) = receipt.TenantID
		*targets[2].(*uint64) = receipt.CutoverGeneration
		*targets[3].(*uint64) = receipt.SourceCount
		*targets[4].(*uint64) = receipt.AdoptedCount
		*targets[5].(*uint64) = receipt.RetiredCount
		*targets[6].(*[]byte) = receipt.RootSHA256[:]
		*targets[7].(*time.Time) = receipt.AcceptedAt
		*targets[8].(*bool) = false
		return nil
	}))
	result, replay, err := store.Accept(context.Background(), testAuthority, receipt)
	if err != nil || replay || result != receipt {
		t.Fatalf("result=%+v replay=%v error=%v", result, replay, err)
	}
}

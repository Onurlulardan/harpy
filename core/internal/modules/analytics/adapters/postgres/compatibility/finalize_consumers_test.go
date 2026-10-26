package compatibility

import (
	"context"
	"testing"
)

func TestFinalizeConsumersBindsInventoryDigestToExactMapping(t *testing.T) {
	mapping := legacyMapping()
	finalized := mapping
	finalized.ConsumersAckedAt = &testTime
	store, database := newStore(t, mappingRow(finalized))
	result, err := store.FinalizeConsumers(
		context.Background(), mapping, [32]byte{3}, testTime,
	)
	if err != nil || result.ConsumersAckedAt == nil ||
		!result.ConsumersAckedAt.Equal(testTime) || len(database.args) != 10 {
		t.Fatalf("result=%+v args=%v error=%v", result, database.args, err)
	}
}

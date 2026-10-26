package compatibility

import (
	"context"
	"testing"
)

func TestRecordHitBindsWholeImmutableMappingIdentity(t *testing.T) {
	mapping := legacyMapping()
	hit := mapping
	hit.HitCount++
	hit.LastHitAt = &testTime
	store, database := newStore(t, mappingRow(hit))
	result, err := store.RecordHit(context.Background(), mapping, testTime)
	if err != nil || result.HitCount != hit.HitCount || result.LastHitAt == nil ||
		len(database.args) != 9 {
		t.Fatalf("result=%+v args=%v error=%v", result, database.args, err)
	}
}

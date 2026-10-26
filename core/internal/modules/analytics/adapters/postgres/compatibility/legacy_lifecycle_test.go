package compatibility

import (
	"context"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func TestLegacyLifecycleBindsWholeImmutableMappingIdentity(t *testing.T) {
	mapping := legacyMapping()
	store, database := newStore(t, mappingRow(mapping))
	resolved, err := store.Resolve(context.Background(), mapping.Key)
	if err != nil || resolved.Key != mapping.Key || len(database.args) != 4 {
		t.Fatalf("resolved=%+v args=%v error=%v", resolved, database.args, err)
	}

	hit := mapping
	hit.HitCount++
	hit.LastHitAt = &testTime
	database.row = mappingRow(hit)
	if _, err = store.RecordHit(context.Background(), mapping, testTime); err != nil ||
		len(database.args) != 9 {
		t.Fatalf("record hit args=%v error=%v", database.args, err)
	}

	acked := mapping
	acked.ConsumersAckedAt = &testTime
	database.row = mappingRow(acked)
	if _, err = store.FinalizeConsumers(context.Background(), mapping, [32]byte{3}, testTime); err != nil ||
		len(database.args) != 10 {
		t.Fatalf("ack args=%v error=%v", database.args, err)
	}

	retired := mapping
	retired.State = "retired"
	database.row = mappingRow(retired)
	evidence := compatibilityapp.RetirementEvidence{AllTenantsV2: true,
		SHA256: [32]byte{4}, ObservedAt: testTime.Add(-time.Minute)}
	if _, err = store.Retire(context.Background(), mapping, evidence, testTime); err != nil ||
		len(database.args) != 11 {
		t.Fatalf("retire args=%v error=%v", database.args, err)
	}
}

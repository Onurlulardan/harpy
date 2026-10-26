package compatibility

import (
	"context"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
)

func TestRetireBindsFreshAuthorityEvidenceAndWholeMapping(t *testing.T) {
	mapping := legacyMapping()
	retired := mapping
	retired.State = "retired"
	store, database := newStore(t, mappingRow(retired))
	evidence := compatibilityapp.RetirementEvidence{AllTenantsV2: true,
		SHA256: [32]byte{4}, ObservedAt: testTime.Add(-time.Minute)}
	result, err := store.Retire(context.Background(), mapping, evidence, testTime)
	if err != nil || result.State != "retired" || len(database.args) != 11 {
		t.Fatalf("result=%+v args=%v error=%v", result, database.args, err)
	}
}

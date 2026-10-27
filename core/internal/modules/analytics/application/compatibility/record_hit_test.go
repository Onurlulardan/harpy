package compatibility

import (
	"context"
	"testing"
)

func TestRecordHitBindsExpectedMappingIdentity(t *testing.T) {
	mapping := mappingFixture()
	application, _ := applicationFixture(t, mapping)
	result, err := application.RecordHit(context.Background(), mapping)
	if err != nil || result.HitCount != mapping.HitCount+1 ||
		result.LastHitAt == nil || *result.LastHitAt != compatNow {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

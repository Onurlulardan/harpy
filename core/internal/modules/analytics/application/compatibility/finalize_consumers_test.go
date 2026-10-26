package compatibility

import (
	"context"
	"testing"
)

func TestFinalizeConsumersRequiresInventoryDigest(t *testing.T) {
	mapping := mappingFixture()
	mapping.ConsumersAckedAt = nil
	application, _ := applicationFixture(t, mapping)
	if _, err := application.FinalizeConsumers(context.Background(), mapping, [32]byte{}); err != ErrInvalid {
		t.Fatalf("empty digest error=%v", err)
	}
	result, err := application.FinalizeConsumers(context.Background(), mapping, [32]byte{4})
	if err != nil || result.ConsumersAckedAt == nil || *result.ConsumersAckedAt != compatNow {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

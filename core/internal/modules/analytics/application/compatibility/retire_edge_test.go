package compatibility

import (
	"context"
	"testing"
	"time"
)

func TestRetireEdgeRequiresThirtyDaySilenceAndCurrentV2Evidence(t *testing.T) {
	mapping := mappingFixture()
	application, _ := applicationFixture(t, mapping)
	result, err := application.RetireEdge(context.Background(), mapping)
	if err != nil || result.State != "retired" {
		t.Fatalf("result=%+v error=%v", result, err)
	}

	mapping.LastHitAt = timePointer(compatNow.Add(-29 * 24 * time.Hour))
	application, _ = applicationFixture(t, mapping)
	if _, err = application.RetireEdge(context.Background(), mapping); err != ErrBlocked {
		t.Fatalf("silence error=%v", err)
	}

	mapping.LastHitAt = nil
	application, _ = applicationFixture(t, mapping)
	if _, err = application.RetireEdge(context.Background(), mapping); err != nil {
		t.Fatalf("never-hit retirement error=%v", err)
	}
}

func timePointer(value time.Time) *time.Time { return &value }

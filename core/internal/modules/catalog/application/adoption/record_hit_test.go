package adoption

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRecordHitRejectsEmptyRoute(t *testing.T) {
	err := (&Application{}).RecordHit(context.Background(), platformtx.TenantScope{}, RouteRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("record hit error = %v", err)
	}
}

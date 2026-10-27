package adoption

import (
	"context"
	"errors"
	"testing"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

func TestRecordHitRejectsEmptyRoute(t *testing.T) {
	err := (&store{}).RecordHit(context.Background(), adoptionapp.RouteRequest{}, time.Time{})
	if !errors.Is(err, adoptionapp.ErrInvalid) {
		t.Fatalf("record compatibility hit error = %v", err)
	}
}

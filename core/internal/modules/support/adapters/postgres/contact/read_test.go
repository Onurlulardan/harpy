package contact

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestReadContactStatsFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).ReadContactStats(context.Background(), ports.Access{}, time.Time{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("read contact stats error = %v, want unavailable", err)
	}
}

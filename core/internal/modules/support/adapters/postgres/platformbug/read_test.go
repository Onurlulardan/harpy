package platformbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestReadPlatformBugStatsFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).ReadPlatformBugStats(context.Background(), ports.Access{}, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("read platform stats error = %v, want unavailable", err)
	}
}

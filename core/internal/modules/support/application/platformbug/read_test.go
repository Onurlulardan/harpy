package platformbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestReadPlatformBugStatsFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).ReadPlatformBugStats(context.Background(), ports.Access{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("read platform stats error = %v, want unavailable", err)
	}
}

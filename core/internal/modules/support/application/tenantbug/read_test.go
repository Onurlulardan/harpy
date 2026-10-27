package tenantbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestReadTenantBugStatsFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).ReadTenantBugStats(context.Background(), ports.Access{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("read tenant stats error = %v, want unavailable", err)
	}
}

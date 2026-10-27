package tenantbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestReadTenantBugStatsFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).ReadTenantBugStats(context.Background(), ports.Access{}, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("read tenant stats error = %v, want unavailable", err)
	}
}

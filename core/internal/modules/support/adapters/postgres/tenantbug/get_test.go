package tenantbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestGetTenantBugFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).GetTenantBug(context.Background(), ports.Access{}, domain.NilID, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("get tenant bug error = %v, want unavailable", err)
	}
}

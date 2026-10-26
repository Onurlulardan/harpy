package tenantbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestUpdateTenantBugFailsClosedWithoutDriver(t *testing.T) {
	_, _, err := (*Store)(nil).UpdateTenantBug(context.Background(), ports.Access{}, domain.BugUpdate{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("update tenant bug error = %v, want unavailable", err)
	}
}

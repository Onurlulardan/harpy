package tenantbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestGetTenantBugFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).GetTenantBug(context.Background(), ports.Access{}, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("get tenant bug error = %v, want unavailable", err)
	}
}

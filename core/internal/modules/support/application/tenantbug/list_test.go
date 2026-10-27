package tenantbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestListTenantBugsFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).ListTenantBugs(context.Background(), ports.Access{}, domain.BugFilter{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("list tenant bugs error = %v, want unavailable", err)
	}
}

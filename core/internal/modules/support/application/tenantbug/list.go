package tenantbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) ListTenantBugs(ctx context.Context, access ports.Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	if !service.ready() {
		return domain.BugPage{}, ports.ErrUnavailable
	}
	return service.kernel.TenantBugsList(ctx, access, filter)
}

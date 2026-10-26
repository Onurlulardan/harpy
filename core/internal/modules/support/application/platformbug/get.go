package platformbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) GetPlatformBug(ctx context.Context, access ports.Access,
	id, tenantID domain.ID,
) (domain.BugReport, error) {
	if !service.ready() {
		return domain.BugReport{}, ports.ErrUnavailable
	}
	return service.kernel.PlatformBugGet(ctx, access, id, tenantID)
}

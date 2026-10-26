package platformbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) UpdatePlatformBug(ctx context.Context, access ports.Access,
	update domain.BugUpdate,
) (domain.BugReport, error) {
	if !service.ready() {
		return domain.BugReport{}, ports.ErrUnavailable
	}
	return service.kernel.PlatformBugUpdate(ctx, access, update)
}

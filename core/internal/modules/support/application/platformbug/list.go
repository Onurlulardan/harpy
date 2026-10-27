package platformbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) ListPlatformBugs(ctx context.Context, access ports.Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	if !service.ready() {
		return domain.BugPage{}, ports.ErrUnavailable
	}
	return service.kernel.PlatformBugsList(ctx, access, filter)
}

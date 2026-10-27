package platformbug

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func (service *Service) ReadPlatformBugStats(ctx context.Context,
	access ports.Access,
) (domain.BugStats, error) {
	if !service.ready() {
		return domain.BugStats{}, ports.ErrUnavailable
	}
	return service.kernel.PlatformBugStatsRead(ctx, access)
}

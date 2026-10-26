package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) PlatformBugStatsRead(ctx context.Context,
	access Access,
) (domain.BugStats, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformBugRead, OperationBugPlatformStats) {
		return domain.BugStats{}, ErrForbidden
	}
	stats, err := service.store.BugStatsRead(ctx, access, domain.NilID)
	return stats, normalizeStoreError(err)
}

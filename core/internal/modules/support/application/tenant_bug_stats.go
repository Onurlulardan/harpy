package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) TenantBugStatsRead(ctx context.Context,
	access Access,
) (domain.BugStats, error) {
	if !access.Allows(ScopeTenant, PermissionBugRead, OperationBugTenantStats) {
		return domain.BugStats{}, ErrForbidden
	}
	stats, err := service.store.BugStatsRead(ctx, access, access.TenantID())
	return stats, normalizeStoreError(err)
}

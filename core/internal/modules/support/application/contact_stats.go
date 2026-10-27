package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ContactStatsRead(ctx context.Context,
	access Access,
) (domain.ContactStats, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformContactRead, OperationContactStats) {
		return domain.ContactStats{}, ErrForbidden
	}
	now, err := service.now()
	if err != nil {
		return domain.ContactStats{}, err
	}
	stats, err := service.store.ContactStatsRead(ctx, access, now)
	return stats, normalizeStoreError(err)
}

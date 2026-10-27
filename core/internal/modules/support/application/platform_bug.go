package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) PlatformBugGet(ctx context.Context, access Access,
	id, tenantID domain.ID,
) (domain.BugReport, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformBugRead, OperationBugPlatformRead) {
		return domain.BugReport{}, ErrForbidden
	}
	if id == domain.NilID || tenantID == domain.NilID {
		return domain.BugReport{}, ErrInvalid
	}
	report, err := service.store.BugGet(ctx, access, id, tenantID)
	return report, normalizeStoreError(err)
}

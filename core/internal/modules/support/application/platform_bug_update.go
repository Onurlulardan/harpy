package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) PlatformBugUpdate(ctx context.Context, access Access,
	update domain.BugUpdate,
) (domain.BugReport, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformBugManage, OperationBugPlatformUpdate) {
		return domain.BugReport{}, ErrForbidden
	}
	if update.TenantID == domain.NilID {
		return domain.BugReport{}, ErrInvalid
	}
	return service.updateBug(ctx, access, update)
}

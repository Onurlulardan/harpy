package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) TenantBugUpdate(ctx context.Context, access Access,
	update domain.BugUpdate,
) (domain.BugReport, error) {
	if !access.Allows(ScopeTenant, PermissionBugManage, OperationBugTenantUpdate) {
		return domain.BugReport{}, ErrForbidden
	}
	update.TenantID = domain.ID(access.TenantID())
	return service.updateBug(ctx, access, update)
}

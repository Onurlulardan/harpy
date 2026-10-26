package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) TenantBugsList(ctx context.Context, access Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	if !access.Allows(ScopeTenant, PermissionBugRead, OperationBugTenantList) {
		return domain.BugPage{}, ErrForbidden
	}
	filter.TenantID, filter.ReporterID = domain.ID(access.TenantID()), domain.NilID
	return service.listBugs(ctx, access, filter, false)
}

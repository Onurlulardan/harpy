package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ReporterBugsList(ctx context.Context, access Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	if !access.Allows(ScopeReporter, PermissionBugReadOwn, OperationBugMine) {
		return domain.BugPage{}, ErrForbidden
	}
	filter.TenantID, filter.ReporterID = domain.ID(access.TenantID()), domain.ID(access.ActorID())
	return service.listBugs(ctx, access, filter, true)
}

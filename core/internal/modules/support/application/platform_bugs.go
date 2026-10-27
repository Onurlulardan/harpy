package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) PlatformBugsList(ctx context.Context, access Access,
	filter domain.BugFilter,
) (domain.BugPage, error) {
	if !access.Allows(ScopePlatform, PermissionPlatformBugRead, OperationBugPlatformList) {
		return domain.BugPage{}, ErrForbidden
	}
	filter.ReporterID = domain.NilID
	return service.listBugs(ctx, access, filter, false)
}

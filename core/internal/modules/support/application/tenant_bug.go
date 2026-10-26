package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) TenantBugGet(ctx context.Context, access Access,
	id domain.ID,
) (domain.BugReport, error) {
	if !access.Allows(ScopeTenant, PermissionBugRead, OperationBugTenantRead) {
		return domain.BugReport{}, ErrForbidden
	}
	if id == domain.NilID {
		return domain.BugReport{}, ErrInvalid
	}
	report, err := service.store.BugGet(ctx, access, id, access.TenantID())
	return report, normalizeStoreError(err)
}

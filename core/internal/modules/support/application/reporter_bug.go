package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func (service *Service) ReporterBugGet(ctx context.Context, access Access,
	id domain.ID,
) (domain.BugReport, error) {
	if !access.Allows(ScopeReporter, PermissionBugReadOwn, OperationBugRead) {
		return domain.BugReport{}, ErrForbidden
	}
	if id == domain.NilID {
		return domain.BugReport{}, ErrInvalid
	}
	report, err := service.store.BugGet(ctx, access, id, access.TenantID())
	if err == nil && report.ReporterID != domain.ID(access.ActorID()) {
		return domain.BugReport{}, ErrNotFound
	}
	return reporterView(report), normalizeStoreError(err)
}

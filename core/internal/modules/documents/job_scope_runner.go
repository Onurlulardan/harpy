package documents

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type JobScopeCoordinator interface {
	InRepeatableReadJobTenant(context.Context, uuid.UUID,
		func(context.Context, platformtx.JobTenantScope) error) error
}

type jobScopeRunner struct{ coordinator JobScopeCoordinator }

func (runner jobScopeRunner) InRepeatableReadJobTenant(ctx context.Context,
	tenantID foundation.ID,
	execute func(context.Context, platformtx.JobTenantScope) error,
) error {
	if runner.coordinator == nil || ctx == nil || ctx.Err() != nil ||
		!tenantID.Valid() || execute == nil {
		return ErrDocumentsRuntime
	}
	return runner.coordinator.InRepeatableReadJobTenant(ctx, uuid.UUID(tenantID), execute)
}

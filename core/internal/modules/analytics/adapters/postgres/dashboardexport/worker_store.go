package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type WorkerStore struct {
	source      platformasync.AnalyticsDashboardExportSource
	runInTenant func(context.Context, uuid.UUID, func(context.Context, *workerBound) error) error
}

type workerBound struct {
	queryRow      func(context.Context, string, ...any) row
	currentSchema func(context.Context) (string, error)
}

func (storage *WorkerStore) inTenant(
	ctx context.Context, tenantID uuid.UUID,
	execute func(context.Context, *workerBound) error,
) error {
	if storage == nil || storage.runInTenant == nil || ctx == nil || ctx.Err() != nil ||
		tenantID == uuid.Nil || execute == nil {
		return exportapp.ErrUnavailable
	}
	return storage.runInTenant(ctx, tenantID, execute)
}

var _ exportapp.WorkerStore = (*WorkerStore)(nil)
var _ exportapp.RenderClaimStore = (*WorkerStore)(nil)
var _ exportapp.RecoveryStore = (*WorkerStore)(nil)

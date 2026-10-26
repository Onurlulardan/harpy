package dashboardexport

import (
	"context"
	"fmt"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	analyticsworker "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres/analyticsdashboardexport"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func BindWorker(
	pool *platformpostgres.Pool, coordinator *platformtx.Coordinator,
) (*WorkerStore, error) {
	if coordinator == nil {
		return nil, fmt.Errorf("bind analytics dashboard export worker: coordinator unavailable")
	}
	source, err := analyticsworker.Bind(pool)
	if err != nil {
		return nil, fmt.Errorf("bind analytics dashboard export worker: %w", err)
	}
	worker := &WorkerStore{source: source}
	owner, err := platformtx.NewJobOwnerBinder(
		storageOwner, func(database platformtx.JobBoundExecutor) *workerBound {
			return &workerBound{queryRow: func(
				ctx context.Context, statement string, arguments ...any,
			) row {
				return database.QueryRow(ctx, statement, arguments...)
			}, currentSchema: database.CurrentSchema}
		})
	if err != nil {
		return nil, fmt.Errorf("bind analytics dashboard export owner: %w", err)
	}
	worker.runInTenant = func(ctx context.Context, tenantID uuid.UUID,
		execute func(context.Context, *workerBound) error,
	) error {
		return coordinator.InJobTenant(ctx, tenantID,
			func(boundContext context.Context, scope platformtx.JobTenantScope) error {
				bound, bindErr := platformtx.BindJob(scope, owner)
				if bindErr != nil || bound == nil {
					return exportapp.ErrUnavailable
				}
				return execute(boundContext, bound)
			})
	}
	return worker, nil
}

func (binder *Binder) Bind(
	ctx context.Context, scope platformtx.TenantScope, access exportapp.ScopeAccess,
) (exportapp.Store, exportapp.ScopeIdentity, error) {
	mode, valid := exportAccessMode(access)
	identity := exportapp.ScopeIdentity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if binder == nil || !valid || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!exportScopeUsable(ctx, scope, mode) {
		return nil, exportapp.ScopeIdentity{}, exportapp.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: identity.TenantID, actorID: identity.ActorID,
				command: mode == platformtx.Command,
				queryRow: func(call context.Context, statement string, arguments ...any) row {
					return database.QueryRow(call, statement, arguments...)
				}, allowed: func(call context.Context) bool {
					return exportScopeUsable(call, scope, mode)
				}, notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, exportapp.ScopeIdentity{}, exportapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, exportapp.ScopeIdentity{}, exportapp.ErrUnavailable
	}
	return bound, identity, nil
}

func exportAccessMode(access exportapp.ScopeAccess) (platformtx.ScopeMode, bool) {
	if access == exportapp.ScopeCommand {
		return platformtx.Command, true
	}
	if access == exportapp.ScopeSnapshot {
		return platformtx.SnapshotQuery, true
	}
	return "", false
}

func exportScopeUsable(ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ exportapp.StoreBinder = (*Binder)(nil)

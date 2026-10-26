package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}
type tenantScope = platformtx.TenantScope
type scopeMode = platformtx.ScopeMode
type boundDatabase = platformpostgres.BoundExecutor

const (
	commandScopeMode       scopeMode = platformtx.Command
	snapshotQueryScopeMode scopeMode = platformtx.SnapshotQuery
)

func NewBinder() *Binder { return &Binder{} }

func NewAttemptStore(coordinator *platformtx.Coordinator) (*AttemptStore, error) {
	if coordinator == nil {
		return nil, app.ErrDependencies
	}
	binder, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *attemptBound {
			return &attemptBound{database: database}
		})
	if err != nil {
		return nil, app.ErrDependencies
	}
	store := &AttemptStore{}
	store.runInTenant = func(ctx context.Context, tenantID uuid.UUID,
		execute func(context.Context, *attemptBound) error,
	) error {
		return coordinator.InJobTenant(ctx, tenantID,
			func(scoped context.Context, scope platformtx.JobTenantScope) error {
				bound, bindErr := platformtx.BindJob(scope, binder)
				if bindErr != nil || bound == nil || platformtx.NetworkAllowed(scoped) ||
					!platformtx.ContextOwnsJobScope(scoped, scope) {
					return app.ErrWorkerUnavailable
				}
				return execute(scoped, bound)
			})
	}
	return store, nil
}

func NewWorkerStore() (*WorkerStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *workerBound {
			return &workerBound{database: database,
				notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, app.ErrDependencies
	}
	store := &WorkerStore{}
	store.bindOwned = func(scope platformtx.TenantScope) (*workerBound, error) {
		bound, bindErr := platformtx.Bind(scope, binder)
		if bindErr != nil || bound == nil {
			return nil, app.ErrAuthorizationCheckpoint
		}
		return bound, nil
	}
	return store, nil
}

func NewOutboundAttemptStore(
	coordinator *platformtx.Coordinator,
) (*OutboundAttemptStore, error) {
	if coordinator == nil {
		return nil, app.ErrExportUnavailable
	}
	binder, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *exportAttemptBound {
			return &exportAttemptBound{database: database}
		})
	if err != nil {
		return nil, app.ErrExportUnavailable
	}
	store := &OutboundAttemptStore{}
	store.runInTenant = func(ctx context.Context, tenantID uuid.UUID,
		execute func(context.Context, *exportAttemptBound) error,
	) error {
		return coordinator.InJobTenant(ctx, tenantID,
			func(scoped context.Context, scope platformtx.JobTenantScope) error {
				bound, bindErr := platformtx.BindJob(scope, binder)
				if bindErr != nil || bound == nil || platformtx.NetworkAllowed(scoped) ||
					!platformtx.ContextOwnsJobScope(scoped, scope) {
					return app.ErrExportUnavailable
				}
				return execute(scoped, bound)
			})
	}
	return store, nil
}

func NewReconciliationStore(
	coordinator *platformtx.Coordinator,
) (*ExportReconciliationStore, error) {
	if coordinator == nil {
		return nil, app.ErrExportUnavailable
	}
	binder, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *exportReconciliationBound {
			return &exportReconciliationBound{database: database}
		})
	if err != nil {
		return nil, app.ErrExportUnavailable
	}
	store := &ExportReconciliationStore{}
	store.runInTenant = func(ctx context.Context, tenantID uuid.UUID,
		execute func(context.Context, *exportReconciliationBound) error,
	) error {
		return coordinator.InJobTenant(ctx, tenantID,
			func(scoped context.Context, scope platformtx.JobTenantScope) error {
				bound, bindErr := platformtx.BindJob(scope, binder)
				if bindErr != nil || bound == nil || platformtx.NetworkAllowed(scoped) ||
					!platformtx.ContextOwnsJobScope(scoped, scope) {
					return app.ErrExportUnavailable
				}
				return execute(scoped, bound)
			})
	}
	return store, nil
}

func NewOutboundStore() (*OutboundStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) boundDatabase { return database })
	if err != nil {
		return nil, app.ErrExportUnavailable
	}
	store := &OutboundStore{}
	store.bindOwned = func(scope tenantScope) (boundDatabase, error) {
		bound, bindErr := platformtx.Bind(scope, binder)
		if bindErr != nil {
			return boundDatabase{}, app.ErrExportUnavailable
		}
		return bound, nil
	}
	return store, nil
}

func (binder *Binder) BindControl(ctx context.Context, scope platformtx.TenantScope,
	access app.ControlScopeAccess) (app.ControlStore, error) {
	mode, ok := controlAccessMode(access, scope)
	if binder == nil || !ok || !controlScopeUsable(ctx, scope, mode) {
		return nil, app.ErrControlUnavailable
	}
	owner, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: scope.TenantID(), actorID: scope.ActorID(),
				command: mode == commandScopeMode,
				queryRow: func(ctx context.Context, sql string, arguments ...any) row {
					return database.QueryRow(ctx, sql, arguments...)
				}, query: func(ctx context.Context, sql string, arguments ...any) (rows, error) {
					return database.Query(ctx, sql, arguments...)
				}, allowed: func(ctx context.Context) bool {
					return controlScopeUsable(ctx, scope, mode)
				}, notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, app.ErrControlUnavailable
	}
	bound, err := platformtx.Bind(scope, owner)
	if err != nil || bound == nil {
		return nil, app.ErrControlUnavailable
	}
	return bound, nil
}

func controlScopeUsable(ctx context.Context, scope platformtx.TenantScope,
	mode platformtx.ScopeMode) bool {
	return ownedScopeUsable(ctx, scope, mode)
}

func ownedScopeUsable(ctx context.Context, scope tenantScope, mode scopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, mode) == nil
}

var _ app.ControlStoreBinder = (*Binder)(nil)

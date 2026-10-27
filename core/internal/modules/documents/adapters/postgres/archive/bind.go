package archive

import (
	"context"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context, scope platformtx.TenantScope,
	access archiveapp.Access,
) (archiveapp.Store, foundation.Identity, error) {
	mode, ok := archiveAccessMode(access)
	actorID, actorErr := domain.ParseID(scope.ActorID().String())
	tenantID, tenantErr := domain.ParseID(scope.TenantID().String())
	identity := foundation.Identity{ActorID: actorID, TenantID: tenantID}
	if binder == nil || !ok || actorErr != nil || tenantErr != nil ||
		!archiveScopeUsable(ctx, scope, mode) {
		return nil, foundation.Identity{}, archiveapp.ErrUnavailable
	}
	owner, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return boundStore(database, actorID, tenantID, mode == platformtx.Command,
				false, func(ctx context.Context) bool {
					return archiveScopeUsable(ctx, scope, mode)
				})
		})
	if err != nil {
		return nil, foundation.Identity{}, archiveapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, owner)
	if err != nil || bound == nil {
		return nil, foundation.Identity{}, archiveapp.ErrUnavailable
	}
	return bound, identity, nil
}

func archiveAccessMode(access archiveapp.Access) (platformtx.ScopeMode, bool) {
	switch access {
	case archiveapp.Command:
		return platformtx.Command, true
	case archiveapp.SnapshotAccess:
		return platformtx.SnapshotQuery, true
	default:
		return "", false
	}
}

func archiveScopeUsable(ctx context.Context, scope platformtx.TenantScope,
	mode platformtx.ScopeMode,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

func (binder *Binder) BindJob(ctx context.Context,
	scope platformtx.JobTenantScope,
) (archiveapp.Store, error) {
	tenantID, err := domain.ParseID(scope.TenantID().String())
	if binder == nil || err != nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsJobScope(ctx, scope) {
		return nil, archiveapp.ErrUnavailable
	}
	owner, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return boundStore(database, domain.ID{}, tenantID, true, true,
				func(value context.Context) bool {
					return !platformtx.NetworkAllowed(value) &&
						platformtx.ContextOwnsJobScope(value, scope)
				})
		})
	if err != nil {
		return nil, archiveapp.ErrUnavailable
	}
	bound, err := platformtx.BindJob(scope, owner)
	if err != nil || bound == nil {
		return nil, archiveapp.ErrUnavailable
	}
	return bound, nil
}

func NewJobStore(database platformtx.JobBoundExecutor,
	tenantID domain.ID,
) (archiveapp.Store, error) {
	if !tenantID.Valid() {
		return nil, archiveapp.ErrInvalid
	}
	return boundStore(database, domain.ID{}, tenantID, true, true,
		func(ctx context.Context) bool {
			return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx)
		}), nil
}

func boundStore(database platformpostgres.BoundExecutor, actorID, tenantID domain.ID,
	command, job bool, allowed scopeCheck,
) *store {
	return &store{actorID: actorID, tenantID: tenantID, command: command, job: job,
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return database.QueryRow(ctx, statement, arguments...)
		}, query: func(ctx context.Context, statement string, arguments ...any) (rows, error) {
			return database.Query(ctx, statement, arguments...)
		}, execute: func(ctx context.Context, statement string, arguments ...any) (int64, error) {
			result, err := database.Execute(ctx, statement, arguments...)
			return result.RowsAffected(), err
		}, notFound: platformpostgres.RowNotFound, allowed: allowed}
}

var _ archiveapp.StoreBinder = (*Binder)(nil)

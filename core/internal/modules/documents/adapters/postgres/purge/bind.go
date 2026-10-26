package purge

import (
	"context"

	purgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/purge"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) BindJob(ctx context.Context,
	scope platformtx.JobTenantScope,
) (purgeapp.Store, error) {
	tenantID, err := domain.ParseID(scope.TenantID().String())
	if binder == nil || err != nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsJobScope(ctx, scope) {
		return nil, purgeapp.ErrUnavailable
	}
	owner, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: tenantID,
				queryRow: func(value context.Context, sql string, args ...any) row {
					return database.QueryRow(value, sql, args...)
				}, query: func(value context.Context, sql string, args ...any) (rows, error) {
					return database.Query(value, sql, args...)
				}, notFound: platformpostgres.RowNotFound,
				allowed: func(value context.Context) bool {
					return !platformtx.NetworkAllowed(value) &&
						platformtx.ContextOwnsJobScope(value, scope)
				}}
		})
	if err != nil {
		return nil, purgeapp.ErrUnavailable
	}
	bound, err := platformtx.BindJob(scope, owner)
	if err != nil || bound == nil {
		return nil, purgeapp.ErrUnavailable
	}
	return bound, nil
}

func NewJobStore(database platformtx.JobBoundExecutor,
	tenantID domain.ID,
) (purgeapp.Store, error) {
	if !tenantID.Valid() {
		return nil, purgeapp.ErrInvalid
	}
	return &store{tenantID: tenantID,
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return database.QueryRow(ctx, statement, arguments...)
		},
		query: func(ctx context.Context, statement string, arguments ...any) (rows, error) {
			return database.Query(ctx, statement, arguments...)
		},
		notFound: platformpostgres.RowNotFound,
		allowed: func(ctx context.Context) bool {
			return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx)
		},
	}, nil
}

var _ purgeapp.StoreBinder = (*Binder)(nil)

package deletion

import (
	"context"

	deletionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/deletion"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) BindJob(ctx context.Context,
	scope platformtx.JobTenantScope,
) (deletionapp.Store, error) {
	tenantID, err := domain.ParseID(scope.TenantID().String())
	if binder == nil || err != nil || !scopeUsable(ctx, scope) {
		return nil, deletionapp.ErrUnavailable
	}
	ownerBinder, err := platformtx.NewJobOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: tenantID,
				queryRow: func(ctx context.Context, statement string, args ...any) row {
					return database.QueryRow(ctx, statement, args...)
				}, notFound: platformpostgres.RowNotFound,
				allowed: func(ctx context.Context) bool { return scopeUsable(ctx, scope) }}
		})
	if err != nil {
		return nil, deletionapp.ErrUnavailable
	}
	bound, err := platformtx.BindJob(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, deletionapp.ErrUnavailable
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope platformtx.JobTenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsJobScope(ctx, scope)
}

var _ deletionapp.StoreBinder = (*Binder)(nil)

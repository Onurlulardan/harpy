package registry

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(
	ctx context.Context, scope platformtx.TenantScope,
) (packageregistry.Reader, error) {
	if binder == nil || !scopeUsable(ctx, scope) {
		return nil, packageregistry.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{queryRow: func(callCtx context.Context, statement string, arguments ...any) row {
				return database.QueryRow(callCtx, statement, arguments...)
			}, query: func(callCtx context.Context, statement string, arguments ...any) (rows, error) {
				return database.Query(callCtx, statement, arguments...)
			}, notFound: platformpostgres.RowNotFound,
				allowed: func(callCtx context.Context) bool { return scopeUsable(callCtx, scope) }}
		})
	if err != nil {
		return nil, packageregistry.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, packageregistry.ErrUnavailable
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, scope.Mode()) == nil
}

var _ packageregistry.Binder = (*Binder)(nil)
var _ packageregistry.Reader = (*store)(nil)

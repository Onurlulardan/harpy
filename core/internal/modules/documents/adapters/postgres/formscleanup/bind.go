package formscleanup

import (
	"context"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context,
	scope platformtx.TenantScope,
) (formscleanupapp.Store, error) {
	if binder == nil || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, formscleanupapp.ErrUnavailable
	}
	owner, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				queryRow: func(value context.Context, statement string, args ...any) row {
					return database.QueryRow(value, statement, args...)
				},
				query: func(value context.Context, statement string, args ...any) (rows, error) {
					return database.Query(value, statement, args...)
				},
				notFound: platformpostgres.RowNotFound,
				allowed: func(value context.Context) bool {
					return !platformtx.NetworkAllowed(value) &&
						platformtx.ContextOwnsScope(value, scope)
				},
			}
		})
	if err != nil {
		return nil, formscleanupapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, owner)
	if err != nil || bound == nil {
		return nil, formscleanupapp.ErrUnavailable
	}
	return bound, nil
}

var _ formscleanupapp.StoreBinder = (*Binder)(nil)

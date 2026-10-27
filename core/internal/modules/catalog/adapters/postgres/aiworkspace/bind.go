package aiworkspace

import (
	"context"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const storageOwner = "catalog"

type Owner struct {
	usable   func(context.Context) bool
	queryRow func(context.Context, string, ...any) row
	query    func(context.Context, string, ...any) (rows, error)
	notFound func(error) bool
}

func New(ctx context.Context, scope platformtx.TenantScope) (*Owner, error) {
	if !scopeUsable(ctx, scope) {
		return nil, catalogai.ErrUnavailable
	}
	binder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *Owner {
			return &Owner{usable: func(call context.Context) bool {
				return scopeUsable(call, scope)
			}, queryRow: func(call context.Context, statement string, arguments ...any) row {
				return database.QueryRow(call, statement, arguments...)
			}, query: func(call context.Context, statement string, arguments ...any) (rows, error) {
				return database.Query(call, statement, arguments...)
			}, notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, catalogai.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, catalogai.ErrUnavailable
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && scope.Mode() == platformtx.Command &&
		platformtx.RequireActive(scope, platformtx.Command) == nil &&
		platformtx.RequireOwner(scope, storageOwner) == nil
}

type row interface{ Scan(...any) error }
type rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

func (owner *Owner) allowed(ctx context.Context) bool {
	return owner != nil && owner.usable != nil && owner.queryRow != nil &&
		owner.query != nil && owner.notFound != nil && ctx != nil &&
		ctx.Err() == nil && owner.usable(ctx)
}

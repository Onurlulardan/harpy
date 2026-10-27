package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type queryRow func(context.Context, string, ...any) platformpostgres.Row
type queryMany func(context.Context, string, ...any) (platformpostgres.Rows, error)
type execute func(context.Context, string, ...any) error
type scopeCheck func(context.Context) bool

type Store struct {
	queryRow queryRow
	query    queryMany
	execute  execute
	allowed  scopeCheck
	tenantID uuid.UUID
}

func (store *Store) usable(ctx context.Context) bool {
	return store != nil && store.queryRow != nil && store.query != nil &&
		store.execute != nil && store.allowed != nil && store.tenantID != uuid.Nil &&
		ctx != nil && ctx.Err() == nil && store.allowed(ctx)
}

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) BindCutover(ctx context.Context, abstract api.Scope) (app.Store, error) {
	scope, ok := abstract.(platformtx.TenantScope)
	if binder == nil || !ok || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, app.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder("notifications",
		func(database platformpostgres.BoundExecutor) *Store {
			return &Store{tenantID: scope.TenantID(), queryRow: database.QueryRow,
				query: database.Query, execute: func(operation context.Context,
					statement string, arguments ...any) error {
					_, executeErr := database.Execute(operation, statement, arguments...)
					return executeErr
				},
				allowed: func(operation context.Context) bool {
					return operation != nil && operation.Err() == nil &&
						!platformtx.NetworkAllowed(operation) &&
						platformtx.ContextOwnsScope(operation, scope) &&
						platformtx.RequireActive(scope, platformtx.Command) == nil
				}}
		})
	if err != nil {
		return nil, app.ErrUnavailable
	}
	store, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || store == nil {
		return nil, app.ErrUnavailable
	}
	return store, nil
}

var _ app.Binder = (*Binder)(nil)

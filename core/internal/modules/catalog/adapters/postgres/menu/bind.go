package menu

import (
	"context"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(
	ctx context.Context, scope platformtx.TenantScope, access menuapp.ScopeAccess,
) (menuapp.Store, menuapp.Identity, error) {
	mode, ok := accessMode(access)
	identity := menuapp.Identity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if binder == nil || !ok || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, menuapp.Identity{}, unavailable()
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: identity.TenantID, actorID: identity.ActorID,
				command: mode == platformtx.Command,
				queryRow: func(callCtx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(callCtx, statement, arguments...)
				}, query: func(callCtx context.Context, statement string, arguments ...any) (rows, error) {
					return database.Query(callCtx, statement, arguments...)
				}, execute: func(callCtx context.Context, statement string, arguments ...any) (int64, error) {
					result, executeErr := database.Execute(callCtx, statement, arguments...)
					return result.RowsAffected(), executeErr
				}, currentSchema: database.CurrentSchema, notFound: platformpostgres.RowNotFound,
				allowed: func(callCtx context.Context) bool { return scopeUsable(callCtx, scope, mode) }}
		})
	if err != nil {
		return nil, menuapp.Identity{}, unavailable()
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, menuapp.Identity{}, unavailable()
	}
	return bound, identity, nil
}

func accessMode(access menuapp.ScopeAccess) (platformtx.ScopeMode, bool) {
	switch access {
	case menuapp.ScopeCommand:
		return platformtx.Command, true
	case menuapp.ScopeSnapshot:
		return platformtx.SnapshotQuery, true
	default:
		return "", false
	}
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ menuapp.StoreBinder = (*Binder)(nil)

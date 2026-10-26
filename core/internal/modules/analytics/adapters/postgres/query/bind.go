package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(
	ctx context.Context, scope platformtx.TenantScope, access queryapp.ScopeAccess,
) (queryapp.Store, queryapp.ScopeIdentity, error) {
	mode, ok := accessMode(access)
	identity := queryapp.ScopeIdentity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if binder == nil || !ok || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, queryapp.ScopeIdentity{}, queryapp.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{tenantID: identity.TenantID, actorID: identity.ActorID,
				command: mode == platformtx.Command, operation: scope.OperationID(),
				queryRow: func(ctx context.Context, sql string, args ...any) row {
					return database.QueryRow(ctx, sql, args...)
				},
				query: func(ctx context.Context, sql string, args ...any) (rows, error) {
					return database.Query(ctx, sql, args...)
				},
				allowed:  func(ctx context.Context) bool { return scopeUsable(ctx, scope, mode) },
				notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, queryapp.ScopeIdentity{}, queryapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, queryapp.ScopeIdentity{}, queryapp.ErrUnavailable
	}
	return bound, identity, nil
}

func accessMode(access queryapp.ScopeAccess) (platformtx.ScopeMode, bool) {
	switch access {
	case queryapp.ScopeCommand:
		return platformtx.Command, true
	case queryapp.ScopeSnapshot:
		return platformtx.SnapshotQuery, true
	default:
		return "", false
	}
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ queryapp.StoreBinder = (*Binder)(nil)

package publication

import (
	"context"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context, scope platformtx.TenantScope,
	access publicationapp.ScopeAccess,
) (publicationapp.Store, publicationapp.ScopeIdentity, error) {
	mode, ok := accessMode(access)
	identity := publicationapp.ScopeIdentity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if binder == nil || !ok || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, publicationapp.ScopeIdentity{}, publicationapp.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{actorID: identity.ActorID, command: mode == platformtx.Command,
				queryRow: func(ctx context.Context, sql string, args ...any) row {
					return database.QueryRow(ctx, sql, args...)
				}, query: func(ctx context.Context, sql string, args ...any) (rows, error) {
					return database.Query(ctx, sql, args...)
				}, allowed: func(ctx context.Context) bool { return scopeUsable(ctx, scope, mode) },
				notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, publicationapp.ScopeIdentity{}, publicationapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, publicationapp.ScopeIdentity{}, publicationapp.ErrUnavailable
	}
	return bound, identity, nil
}

func accessMode(access publicationapp.ScopeAccess) (platformtx.ScopeMode, bool) {
	if access == publicationapp.ScopeCommand {
		return platformtx.Command, true
	}
	if access == publicationapp.ScopeSnapshot {
		return platformtx.SnapshotQuery, true
	}
	return "", false
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ publicationapp.StoreBinder = (*Binder)(nil)

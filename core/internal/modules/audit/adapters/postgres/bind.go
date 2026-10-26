package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const storageOwner = "audit"

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (*Binder) Bind(
	ctx context.Context, scope platformtx.TenantScope, access application.ScopeAccess,
) (application.Store, application.ScopeIdentity, error) {
	mode, ok := accessMode(access)
	identity := application.ScopeIdentity{ActorID: scope.ActorID(), TenantID: scope.TenantID()}
	if !ok || identity.ActorID == uuid.Nil || identity.TenantID == uuid.Nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, application.ScopeIdentity{}, application.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *Store {
			return &Store{scopeAccess: access,
				usable: func(call context.Context) bool { return scopeUsable(call, scope, mode) },
				queryRow: func(call context.Context, statement string, arguments ...any) row {
					return database.QueryRow(call, statement, arguments...)
				},
				query: func(call context.Context, statement string, arguments ...any) (rows, error) {
					return database.Query(call, statement, arguments...)
				},
				notFound: platformpostgres.RowNotFound}
		})
	if err != nil {
		return nil, application.ScopeIdentity{}, application.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, application.ScopeIdentity{}, application.ErrUnavailable
	}
	return bound, identity, nil
}

func accessMode(access application.ScopeAccess) (platformtx.ScopeMode, bool) {
	switch access {
	case application.ScopeCommand:
		return platformtx.Command, true
	case application.ScopeSnapshot:
		return platformtx.SnapshotQuery, true
	default:
		return "", false
	}
}

func scopeUsable(
	ctx context.Context, scope platformtx.TenantScope, mode platformtx.ScopeMode,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && scope.Mode() == mode &&
		platformtx.RequireActive(scope, mode) == nil &&
		platformtx.RequireOwner(scope, storageOwner) == nil
}

var _ application.StoreBinder = (*Binder)(nil)

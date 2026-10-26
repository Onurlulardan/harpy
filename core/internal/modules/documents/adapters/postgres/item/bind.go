package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context, scope platformtx.TenantScope,
	access itemapp.Access,
) (itemapp.Store, foundation.Identity, error) {
	mode, ok := accessMode(access)
	actorID, actorErr := domain.ParseID(scope.ActorID().String())
	tenantID, tenantErr := domain.ParseID(scope.TenantID().String())
	identity := foundation.Identity{ActorID: actorID, TenantID: tenantID}
	if binder == nil || !ok || actorErr != nil || tenantErr != nil ||
		!scopeUsable(ctx, scope, mode) {
		return nil, foundation.Identity{}, itemapp.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{actorID: actorID, tenantID: tenantID,
				command: mode == platformtx.Command,
				queryRow: func(ctx context.Context, statement string, args ...any) row {
					return database.QueryRow(ctx, statement, args...)
				}, query: func(ctx context.Context, statement string, args ...any) (rows, error) {
					return database.Query(ctx, statement, args...)
				}, notFound: platformpostgres.RowNotFound,
				allowed: func(ctx context.Context) bool { return scopeUsable(ctx, scope, mode) }}
		})
	if err != nil {
		return nil, foundation.Identity{}, itemapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, foundation.Identity{}, itemapp.ErrUnavailable
	}
	return bound, identity, nil
}

func accessMode(access itemapp.Access) (platformtx.ScopeMode, bool) {
	switch access {
	case itemapp.Command:
		return platformtx.Command, true
	case itemapp.SnapshotAccess:
		return platformtx.SnapshotQuery, true
	default:
		return "", false
	}
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope,
	mode platformtx.ScopeMode,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

var _ itemapp.StoreBinder = (*Binder)(nil)

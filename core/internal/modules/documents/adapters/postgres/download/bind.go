package download

import (
	"context"

	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context,
	scope platformtx.TenantScope,
) (downloadapp.Store, foundation.Identity, error) {
	actorID, actorErr := domain.ParseID(scope.ActorID().String())
	tenantID, tenantErr := domain.ParseID(scope.TenantID().String())
	identity := foundation.Identity{ActorID: actorID, TenantID: tenantID}
	if binder == nil || actorErr != nil || tenantErr != nil || !scopeUsable(ctx, scope) {
		return nil, foundation.Identity{}, downloadapp.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{actorID: actorID, tenantID: tenantID,
				queryRow: func(ctx context.Context, statement string, args ...any) row {
					return database.QueryRow(ctx, statement, args...)
				}, notFound: platformpostgres.RowNotFound,
				allowed: func(ctx context.Context) bool { return scopeUsable(ctx, scope) }}
		})
	if err != nil {
		return nil, foundation.Identity{}, downloadapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, foundation.Identity{}, downloadapp.ErrUnavailable
	}
	return bound, identity, nil
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}

var _ downloadapp.StoreBinder = (*Binder)(nil)

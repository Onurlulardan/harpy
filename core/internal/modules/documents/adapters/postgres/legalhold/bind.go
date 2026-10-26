package legalhold

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	holdapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/legalhold"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context,
	scope platformtx.TenantScope,
) (holdapp.Store, foundation.Identity, error) {
	actorID, actorErr := domain.ParseID(scope.ActorID().String())
	tenantID, tenantErr := domain.ParseID(scope.TenantID().String())
	identity := foundation.Identity{ActorID: actorID, TenantID: tenantID}
	if binder == nil || actorErr != nil || tenantErr != nil || ctx == nil ||
		ctx.Err() != nil || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil {
		return nil, foundation.Identity{}, holdapp.ErrUnavailable
	}
	owner, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{actorID: actorID, tenantID: tenantID,
				queryRow: func(value context.Context, sql string, args ...any) row {
					return database.QueryRow(value, sql, args...)
				}, notFound: platformpostgres.RowNotFound,
				allowed: func(value context.Context) bool {
					return !platformtx.NetworkAllowed(value) &&
						platformtx.ContextOwnsScope(value, scope)
				}}
		})
	if err != nil {
		return nil, foundation.Identity{}, holdapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, owner)
	if err != nil || bound == nil {
		return nil, foundation.Identity{}, holdapp.ErrUnavailable
	}
	return bound, identity, nil
}

var _ holdapp.StoreBinder = (*Binder)(nil)

package knowledgefile

import (
	"context"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(ctx context.Context,
	scope platformtx.TenantScope,
) (knowledgeapp.LeaseStore, foundation.Identity, error) {
	actorID, actorErr := domain.ParseID(scope.ActorID().String())
	tenantID, tenantErr := domain.ParseID(scope.TenantID().String())
	identity := foundation.Identity{ActorID: actorID, TenantID: tenantID}
	if binder == nil || actorErr != nil || tenantErr != nil || !scopeUsable(ctx, scope) {
		return nil, foundation.Identity{}, knowledgeappError()
	}
	owner, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(database platformpostgres.BoundExecutor) *store {
			return &store{actorID: actorID, tenantID: tenantID,
				queryRow: func(value context.Context, statement string, args ...any) row {
					return database.QueryRow(value, statement, args...)
				}, query: func(value context.Context, statement string, args ...any) (rows, error) {
					return database.Query(value, statement, args...)
				}, notFound: platformpostgres.RowNotFound,
				allowed: func(value context.Context) bool { return scopeUsable(value, scope) }}
		})
	if err != nil {
		return nil, foundation.Identity{}, knowledgeappError()
	}
	bound, err := platformtx.Bind(scope, owner)
	if err != nil || bound == nil {
		return nil, foundation.Identity{}, knowledgeappError()
	}
	return bound, identity, nil
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	mode := scope.Mode()
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireOwner(scope, storageOwner) == nil &&
		(mode == platformtx.Command || mode == platformtx.SnapshotQuery)
}

func knowledgeappError() error { return knowledgeapi.ErrUnavailable }

var _ knowledgeapp.LeaseBinder = (*Binder)(nil)

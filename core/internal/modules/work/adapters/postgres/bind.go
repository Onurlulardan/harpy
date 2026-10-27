package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const storageOwner = "work"

type scopeBridge struct{ scope platformtx.TenantScope }
type Binder struct{}

type ownerAccess struct {
	binder platformpostgres.OwnerBinder[*ownerPort]
}

type ownerPort struct{ database databaseFacade }

func NewBinder() *Binder                                   { return &Binder{} }
func BridgeScope(scope platformtx.TenantScope) ports.Scope { return &scopeBridge{scope: scope} }

func newOwnerAccess() (*ownerAccess, error) {
	binder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(bound platformpostgres.BoundExecutor) *ownerPort {
			return &ownerPort{database: databaseFromBound(bound)}
		})
	if err != nil {
		return nil, application.ErrUnavailable
	}
	return &ownerAccess{binder: binder}, nil
}

func (access *ownerAccess) bind(ctx context.Context, bridge *scopeBridge) (databaseFacade, error) {
	if access == nil || !scopeUsable(ctx, bridge) {
		return databaseFacade{}, application.ErrUnavailable
	}
	bound, err := platformtx.Bind(bridge.scope, access.binder)
	if err != nil || bound == nil {
		return databaseFacade{}, application.ErrUnavailable
	}
	return bound.database, nil
}

func (*Binder) Bind(ctx context.Context, local ports.Scope) (ports.Stores, error) {
	bridge, ok := local.(*scopeBridge)
	if !ok || !scopeUsable(ctx, bridge) {
		return nil, application.ErrUnavailable
	}
	if !bridge.SnapshotQuery() {
		authority, err := platformtx.MintOwnerAuthority(bridge.scope, storageOwner)
		if err != nil || !authority.OwnsScope(bridge.scope) || authority.Domain() != storageOwner ||
			authority.Generation() < 1 || authority.GroupGeneration() < 1 {
			return nil, application.ErrUnavailable
		}
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(storageOwner,
		func(bound platformpostgres.BoundExecutor) *Store {
			return newStore(bridge, databaseFromBound(bound))
		})
	if err != nil {
		return nil, application.ErrUnavailable
	}
	store, err := platformtx.Bind(bridge.scope, ownerBinder)
	if err != nil || store == nil {
		return nil, application.ErrUnavailable
	}
	return store, nil
}

func databaseFromBound(bound platformpostgres.BoundExecutor) databaseFacade {
	return databaseFacade{
		executeFn: func(ctx context.Context, statement string, arguments ...any) (int64, error) {
			result, err := bound.Execute(ctx, statement, arguments...)
			return result.RowsAffected(), err
		},
		query: func(ctx context.Context, statement string, arguments ...any) (rows, error) {
			return bound.Query(ctx, statement, arguments...)
		},
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return bound.QueryRow(ctx, statement, arguments...)
		},
		currentSchema: bound.CurrentSchema,
	}
}

func (bridge *scopeBridge) TenantID() domain.ID {
	if bridge == nil {
		return domain.ID{}
	}
	return domain.ID(bridge.scope.TenantID())
}
func (bridge *scopeBridge) ActorID() domain.ID {
	if bridge == nil {
		return domain.ID{}
	}
	return domain.ID(bridge.scope.ActorID())
}
func (bridge *scopeBridge) OperationID() string {
	if bridge == nil {
		return ""
	}
	return bridge.scope.OperationID()
}
func (bridge *scopeBridge) SnapshotQuery() bool {
	return bridge != nil && bridge.scope.Mode() == platformtx.SnapshotQuery
}
func (bridge *scopeBridge) PlatformTenantScope() platformtx.TenantScope {
	if bridge == nil {
		return platformtx.TenantScope{}
	}
	return bridge.scope
}

func scopeUsable(ctx context.Context, bridge *scopeBridge) bool {
	if ctx == nil || ctx.Err() != nil || bridge == nil || !bridge.TenantID().Valid() ||
		!bridge.ActorID().Valid() || bridge.OperationID() == "" || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, bridge.scope) {
		return false
	}
	mode := platformtx.Command
	if bridge.SnapshotQuery() {
		mode = platformtx.SnapshotQuery
	}
	return bridge.scope.Mode() == mode && platformtx.RequireActive(bridge.scope, mode) == nil &&
		platformtx.RequireOwner(bridge.scope, storageOwner) == nil
}

var _ ports.UnitOfWork = (*Binder)(nil)

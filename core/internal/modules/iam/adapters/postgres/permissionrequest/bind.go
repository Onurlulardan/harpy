package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type permissionRequestDatabase interface {
	Execute(context.Context, string, ...any) (platformpostgres.CommandResult, error)
	Query(context.Context, string, ...any) (platformpostgres.Rows, error)
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type tenantScope = platformtx.TenantScope
type permissionRequestRow = platformpostgres.Row

type boundPermissionRequestStore struct {
	database permissionRequestDatabase
}

type permissionRequestStore struct {
	binder platformpostgres.OwnerBinder[*boundPermissionRequestStore]
	valid  bool
}

func NewPermissionRequestStore() (permissionrequest.Store, error) {
	return newPermissionRequestStore()
}

func NewPermissionRequestDecisionAuditStore() (
	permissionrequest.DecisionAuditStore,
	error,
) {
	return newPermissionRequestStore()
}

func NewPermissionRequestCatalogStore() (
	permissionrequest.CatalogStore,
	error,
) {
	return newPermissionRequestStore()
}

func newPermissionRequestStore() (*permissionRequestStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundPermissionRequestStore {
			return &boundPermissionRequestStore{database: database}
		},
	)
	if err != nil {
		return nil, permissionrequest.ErrUnavailable
	}
	return &permissionRequestStore{binder: binder, valid: true}, nil
}

var _ permissionrequest.Store = (*permissionRequestStore)(nil)
var _ permissionrequest.DecisionAuditStore = (*permissionRequestStore)(nil)
var _ permissionrequest.CatalogStore = (*permissionRequestStore)(nil)

func (store *permissionRequestStore) bind(
	scope platformtx.TenantScope,
) (*boundPermissionRequestStore, error) {
	if store == nil || !store.valid {
		return nil, permissionrequest.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, permissionrequest.ErrUnavailable
	}
	return bound, nil
}

func permissionRequestCommandScopeUsable(
	ctx context.Context,
	scope tenantScope,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

func permissionRequestSnapshotScopeUsable(
	ctx context.Context,
	scope tenantScope,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}

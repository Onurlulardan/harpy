package profilebackfill

import (
	"context"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type boundStore struct {
	database platformpostgres.BoundExecutor
}

func scopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

type tenantScope = platformtx.TenantScope

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
}

func NewStore() (profilebackfillapp.Store, error) {
	return newStore()
}

func NewSourceReader() (profilebackfillapp.SourcePort, error) {
	return newStore()
}

func newStore() (*store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return &store{binder: binder}, nil
}

func (value *store) bind(scope tenantScope) (*boundStore, error) {
	if value == nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, value.binder)
	if err != nil || bound == nil {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return bound, nil
}

func (bound *boundStore) fetchRecoveryInventory(
	ctx context.Context,
) (platformpostgres.Rows, error) {
	return bound.database.Query(ctx, fetchRecoveryInventorySQL)
}

func (bound *boundStore) closeRecoveryInventory(ctx context.Context) error {
	_, err := bound.database.Execute(ctx, closeRecoveryInventorySQL)
	return err
}

var _ profilebackfillapp.Store = (*store)(nil)
var _ profilebackfillapp.SourcePort = (*store)(nil)

func receiptMissing(err error) bool {
	return platformpostgres.RowNotFound(err)
}

package sessionview

import (
	"context"
	"errors"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errRead = errors.New("IAM session view read failed")

func rejectedRead(err error) bool {
	return platformpostgres.RowNotFound(err)
}

type tenantScope = platformtx.TenantScope

type boundStore struct {
	database platformpostgres.BoundExecutor
}
type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (app.Port, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, errRead
	}
	return &store{binder: binder, valid: true}, nil
}

func (store *store) bind(
	ctx context.Context, scope platformtx.TenantScope,
) (*boundStore, error) {
	if store == nil || !store.valid || ctx == nil || ctx.Err() != nil ||
		platformtx.NetworkAllowed(ctx) || !platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil {
		return nil, errRead
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, errRead
	}
	return bound, nil
}

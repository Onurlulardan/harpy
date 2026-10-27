package securityepoch

import (
	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/securityepoch"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
}

func NewStore() (app.Port, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, app.ErrRejected
	}
	return &store{binder: binder}, nil
}

func (storage *store) bind(scope tenantScope) (*boundStore, error) {
	if storage == nil {
		return nil, app.ErrRejected
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, app.ErrRejected
	}
	return bound, nil
}

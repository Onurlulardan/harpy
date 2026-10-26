package tenantsettingspostgres

import (
	"fmt"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const transactionOwner = "tenancy"

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
}

type boundStore struct {
	database    platformpostgres.BoundExecutor
	rowNotFound func(error) bool
}

func New() (settings.Registry, error) {
	binder, err := platformpostgres.NewOwnerBinder(transactionOwner,
		func(executor platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{
				database:    executor,
				rowNotFound: platformpostgres.RowNotFound,
			}
		})
	if err != nil {
		return nil, fmt.Errorf("build tenant settings registry binder: %w", err)
	}
	return &store{binder: binder}, nil
}

func (storage *store) bind(scope settings.Scope) (*boundStore, error) {
	if storage == nil {
		return nil, settings.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil || bound.rowNotFound == nil {
		return nil, settings.ErrUnavailable
	}
	return bound, nil
}

var _ settings.Registry = (*store)(nil)

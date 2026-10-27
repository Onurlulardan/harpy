package executionpostgres

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	executionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/execution"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = transaction.TenantScope

type boundExecutor struct {
	database platformpostgres.BoundExecutor
}

type adapter struct {
	binder platformpostgres.OwnerBinder[boundExecutor]
}

type driver struct {
	scope    tenantScope
	database platformpostgres.BoundExecutor
}

func New() (executionapp.DriverPort, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		catalog.Owner,
		func(database platformpostgres.BoundExecutor) boundExecutor {
			return boundExecutor{database: database}
		},
	)
	if err != nil {
		return nil, catalog.ErrContract
	}
	return &adapter{binder: binder}, nil
}

func (adapter *adapter) bind(scope tenantScope) (boundExecutor, error) {
	if adapter == nil ||
		transaction.RequireActive(scope, transaction.SnapshotQuery) != nil {
		return boundExecutor{}, catalog.ErrContract
	}
	bound, err := transaction.Bind(scope, adapter.binder)
	if err != nil {
		return boundExecutor{}, catalog.ErrContract
	}
	return bound, nil
}

func activeSnapshot(scope tenantScope) bool {
	return transaction.RequireActive(scope, transaction.SnapshotQuery) == nil
}

var _ restexecution.SnapshotDriver = (*driver)(nil)

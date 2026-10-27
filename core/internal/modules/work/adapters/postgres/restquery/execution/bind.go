package executionpostgres

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope

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

func New() (restexecution.OwnerPort, error) {
	binder, err := platformpostgres.NewOwnerBinder(workquery.Owner,
		func(database platformpostgres.BoundExecutor) boundExecutor {
			return boundExecutor{database: database}
		})
	if err != nil {
		return nil, workquery.ErrUnavailable
	}
	return &adapter{binder: binder}, nil
}

func (owner *adapter) bind(scope tenantScope) (boundExecutor, error) {
	if owner == nil || platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		platformtx.RequireOwner(scope, workquery.Owner) != nil {
		return boundExecutor{}, workquery.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, owner.binder)
	if err != nil {
		return boundExecutor{}, workquery.ErrUnavailable
	}
	return bound, nil
}

var _ restexecution.OwnerPort = (*adapter)(nil)
var _ restexecution.SnapshotDriver = (*driver)(nil)

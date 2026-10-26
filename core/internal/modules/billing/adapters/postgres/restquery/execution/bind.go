package executionpostgres

import (
	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/application/restquery/offering"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type globalScope = transaction.GlobalScope

type boundExecutor struct {
	database platformpostgres.BoundExecutor
}

type adapter struct {
	binder platformpostgres.GlobalOwnerBinder[boundExecutor]
}

type driver struct {
	scope    globalScope
	database platformpostgres.BoundExecutor
}

func New() (restexecution.GlobalOwnerPort, error) {
	binder, err := platformpostgres.NewGlobalOwnerBinder(
		offering.Owner,
		func(database platformpostgres.BoundExecutor) boundExecutor {
			return boundExecutor{database: database}
		},
	)
	if err != nil {
		return nil, offering.ErrContract
	}
	return &adapter{binder: binder}, nil
}

func (adapter *adapter) bind(scope globalScope) (boundExecutor, error) {
	if adapter == nil || transaction.RequireGlobalOwner(scope, offering.Owner) != nil {
		return boundExecutor{}, offering.ErrContract
	}
	bound, err := transaction.BindGlobal(scope, adapter.binder)
	if err != nil {
		return boundExecutor{}, offering.ErrContract
	}
	return bound, nil
}

func activeGlobalSnapshot(scope globalScope) bool {
	return transaction.RequireGlobalOwner(scope, offering.Owner) == nil
}

var _ restexecution.GlobalSnapshotDriver = (*driver)(nil)
var _ restexecution.GlobalOwnerPort = (*adapter)(nil)

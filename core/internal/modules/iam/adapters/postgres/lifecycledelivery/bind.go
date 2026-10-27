package lifecycledelivery

import (
	"context"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type deliveryDatabase interface {
	Execute(context.Context, string, ...any) (platformpostgres.CommandResult, error)
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type Binding struct {
	binder platformtx.JobOwnerBinder[lifecycledeliveryapp.Store]
}

func NewBinding() (Binding, error) {
	binder, err := platformtx.NewJobOwnerBinder(
		"iam", func(database platformtx.JobBoundExecutor) lifecycledeliveryapp.Store {
			return &store{database: database}
		},
	)
	if err != nil {
		return Binding{}, err
	}
	return Binding{binder: binder}, nil
}

func (binding Binding) Bind(
	scope platformtx.JobTenantScope,
) (lifecycledeliveryapp.Store, error) {
	return platformtx.BindJob(scope, binding.binder)
}

func rowNotFound(err error) bool { return platformpostgres.RowNotFound(err) }

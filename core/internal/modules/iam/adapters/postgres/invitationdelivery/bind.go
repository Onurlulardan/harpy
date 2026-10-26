package invitationdelivery

import (
	"context"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type deliveryDatabase interface {
	Execute(context.Context, string, ...any) (platformpostgres.CommandResult, error)
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

func rowNotFound(err error) bool { return platformpostgres.RowNotFound(err) }

type Binding struct {
	binder platformtx.JobOwnerBinder[invitationdeliveryapp.Store]
}

func NewBinding() (Binding, error) {
	binder, err := platformtx.NewJobOwnerBinder(
		"iam", func(database platformtx.JobBoundExecutor) invitationdeliveryapp.Store {
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
) (invitationdeliveryapp.Store, error) {
	return platformtx.BindJob(scope, binding.binder)
}

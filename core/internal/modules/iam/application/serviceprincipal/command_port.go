package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CommandPort interface {
	Create(context.Context, transaction.TenantScope, CreateCommand) (MutationResult, error)
	Disable(context.Context, transaction.TenantScope, StateCommand) (MutationResult, error)
	Enable(context.Context, transaction.TenantScope, StateCommand) (MutationResult, error)
	Retire(context.Context, transaction.TenantScope, StateCommand) (MutationResult, error)
	Grant(context.Context, transaction.TenantScope, GrantCommand) (MutationResult, error)
	Revoke(context.Context, transaction.TenantScope, RevokeCommand) (MutationResult, error)
}

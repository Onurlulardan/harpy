package catalogcomponent

import (
	"context"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type StoreBinder interface {
	Bind(context.Context, platformtx.TenantScope) (Store, error)
}

type Store interface {
	Read(context.Context, contract.Component) (contract.Readiness, error)
	Apply(context.Context, contract.Command) (contract.Receipt, error)
	Blockers(context.Context, contract.Command) (contract.BlockerReceipt, error)
	AllowsMenuTarget(context.Context, string, string) (bool, error)
}

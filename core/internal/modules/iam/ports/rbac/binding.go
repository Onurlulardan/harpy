package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type BindingStore[Scope any] interface {
	Grant(context.Context, Scope, rbacdomain.BindingCommand) (rbacdomain.MutationResult, error)
	Get(context.Context, Scope, rbacdomain.GetQuery) (rbacdomain.Binding, error)
	List(context.Context, Scope, rbacdomain.BindingListQuery) ([]rbacdomain.Binding, error)
	Revoke(context.Context, Scope, rbacdomain.BindingCommand) (rbacdomain.MutationResult, error)
}

package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type ActionStore[Scope any] interface {
	Create(context.Context, Scope, rbacdomain.ActionInput) (rbacdomain.MutationResult, error)
	Get(context.Context, Scope, rbacdomain.GetQuery) (rbacdomain.Action, error)
	List(context.Context, Scope, rbacdomain.ListQuery) ([]rbacdomain.Action, error)
	Update(context.Context, Scope, rbacdomain.UpdateCatalog) (rbacdomain.MutationResult, error)
	Retire(context.Context, Scope, rbacdomain.StateCommand) (rbacdomain.MutationResult, error)
}

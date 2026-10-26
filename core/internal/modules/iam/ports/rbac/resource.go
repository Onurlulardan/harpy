package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type ResourceStore[Scope any] interface {
	Create(context.Context, Scope, rbacdomain.ResourceInput) (rbacdomain.MutationResult, error)
	Get(context.Context, Scope, rbacdomain.GetQuery) (rbacdomain.Resource, error)
	List(context.Context, Scope, rbacdomain.ListQuery) ([]rbacdomain.Resource, error)
	Update(context.Context, Scope, rbacdomain.UpdateCatalog) (rbacdomain.MutationResult, error)
	Retire(context.Context, Scope, rbacdomain.StateCommand) (rbacdomain.MutationResult, error)
}

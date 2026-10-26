package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type PermissionStore[Scope any] interface {
	Create(context.Context, Scope, rbacdomain.PermissionInput) (rbacdomain.MutationResult, error)
	Get(context.Context, Scope, rbacdomain.GetQuery) (rbacdomain.Permission, error)
	List(context.Context, Scope, rbacdomain.ListQuery) ([]rbacdomain.Permission, error)
	Retire(context.Context, Scope, rbacdomain.StateCommand) (rbacdomain.MutationResult, error)
}

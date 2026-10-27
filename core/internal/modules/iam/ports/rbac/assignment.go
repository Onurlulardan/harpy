package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type AssignmentStore[Scope any] interface {
	Create(context.Context, Scope, rbacdomain.AssignmentMutation) (rbacdomain.MutationResult, error)
	Get(context.Context, Scope, rbacdomain.AssignmentGetQuery) (rbacdomain.Assignment, error)
	List(context.Context, Scope, rbacdomain.AssignmentListQuery) ([]rbacdomain.Assignment, error)
	Update(context.Context, Scope, rbacdomain.AssignmentMutation) (rbacdomain.MutationResult, error)
	Retire(context.Context, Scope, rbacdomain.AssignmentMutation) (rbacdomain.MutationResult, error)
	Effective(context.Context, Scope, rbacdomain.EffectivePermissionQuery) ([]rbacdomain.EffectivePermission, error)
}

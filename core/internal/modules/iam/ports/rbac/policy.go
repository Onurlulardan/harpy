package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type ProtectedAdminStore[Scope any] interface {
	Get(context.Context, Scope, rbacdomain.GetQuery) (rbacdomain.ProtectedAdminPolicy, error)
	Update(context.Context, Scope, rbacdomain.UpdatePolicyCommand) (rbacdomain.MutationResult, error)
}

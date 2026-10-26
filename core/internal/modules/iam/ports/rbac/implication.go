package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type ImplicationStore[Scope any] interface {
	Create(context.Context, Scope, rbacdomain.ImplicationInput) (rbacdomain.MutationResult, error)
	Get(context.Context, Scope, rbacdomain.GetQuery) (rbacdomain.Implication, error)
	List(context.Context, Scope, rbacdomain.ListQuery) ([]rbacdomain.Implication, error)
	RecordHit(context.Context, Scope, rbacdomain.RecordHitCommand) (rbacdomain.MutationResult, error)
	Retire(context.Context, Scope, rbacdomain.StateCommand) (rbacdomain.MutationResult, error)
}

package action

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Get(ctx context.Context, scope transaction.TenantScope,
	query rbacdomain.GetQuery,
) (rbacdomain.Action, error) {
	if commands == nil || query.Validate() != nil || !queryScope(ctx, scope, query.Evidence) {
		return rbacdomain.Action{}, rbacdomain.ErrInvalidQuery
	}
	return commands.store.Get(ctx, scope, query)
}

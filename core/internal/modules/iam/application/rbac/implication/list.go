package implication

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) List(ctx context.Context, scope transaction.TenantScope,
	query rbacdomain.ListQuery) ([]rbacdomain.Implication, error) {
	if commands == nil || query.Validate() != nil || !queryScope(ctx, scope, query.Evidence) {
		return nil, rbacdomain.ErrInvalidQuery
	}
	return commands.store.List(ctx, scope, query)
}

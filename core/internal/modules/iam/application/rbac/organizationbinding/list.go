package organizationbinding

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) List(ctx context.Context, scope transaction.TenantScope,
	query rbacdomain.BindingListQuery) ([]rbacdomain.Binding, error) {
	if commands == nil || query.Validate() != nil || !queryScope(ctx, scope, query.Evidence) {
		return nil, rbacdomain.ErrInvalidQuery
	}
	return commands.store.List(ctx, scope, query)
}

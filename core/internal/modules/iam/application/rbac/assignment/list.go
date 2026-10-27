package assignment

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) List(
	ctx context.Context,
	scope platformtx.TenantScope,
	query rbacdomain.AssignmentListQuery,
) ([]rbacdomain.Assignment, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScope(ctx, scope, query.Evidence) {
		return nil, rbacdomain.ErrInvalidQuery
	}
	return commands.store.List(ctx, scope, query)
}

package assignment

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Get(
	ctx context.Context,
	scope platformtx.TenantScope,
	query rbacdomain.AssignmentGetQuery,
) (rbacdomain.Assignment, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScope(ctx, scope, query.Evidence) {
		return rbacdomain.Assignment{}, rbacdomain.ErrInvalidQuery
	}
	return commands.store.Get(ctx, scope, query)
}

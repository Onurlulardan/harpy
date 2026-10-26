package assignment

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (commands *Commands) Effective(
	ctx context.Context,
	scope platformtx.TenantScope,
	query rbacdomain.EffectivePermissionQuery,
) ([]rbacdomain.EffectivePermission, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScope(ctx, scope, query.Evidence) {
		return nil, rbacdomain.ErrInvalidQuery
	}
	return commands.store.Effective(ctx, scope, query)
}

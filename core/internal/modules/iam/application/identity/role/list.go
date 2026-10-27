package role

import (
	"context"
)

func (commands *Commands) ListRoles(
	ctx context.Context,
	scope TenantScope,
	query ListRolesQuery,
) ([]Role, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return nil, ErrInvalidQuery
	}
	result, err := commands.store.ListRoles(ctx, scope, query)
	if err != nil || len(result) > query.Limit {
		return nil, errOrStorage(err)
	}
	for _, role := range result {
		if role.Validate() != nil {
			return nil, ErrStorage
		}
	}
	return result, nil
}

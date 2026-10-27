package role

import (
	"context"
)

func (commands *Commands) GetRole(
	ctx context.Context,
	scope TenantScope,
	query GetRoleQuery,
) (Role, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return Role{}, ErrInvalidQuery
	}
	result, err := commands.store.GetRole(ctx, scope, query)
	if err != nil || result.Validate() != nil {
		if err != nil {
			return Role{}, err
		}
		return Role{}, ErrStorage
	}
	return result, nil
}

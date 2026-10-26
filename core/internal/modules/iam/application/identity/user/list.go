package user

import (
	"context"
)

func (commands *Commands) ListUsers(
	ctx context.Context,
	scope TenantScope,
	query ListUsersQuery,
) ([]User, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return nil, ErrInvalidQuery
	}
	result, err := commands.store.ListUsers(ctx, scope, query)
	if err != nil || len(result) > query.Limit {
		return nil, errOrStorage(err)
	}
	for _, user := range result {
		if user.Validate() != nil {
			return nil, ErrStorage
		}
	}
	return result, nil
}

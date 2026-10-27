package user

import (
	"context"
)

func (commands *Commands) GetUser(
	ctx context.Context,
	scope TenantScope,
	query GetUserQuery,
) (User, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return User{}, ErrInvalidQuery
	}
	result, err := commands.store.GetUser(ctx, scope, query)
	if err != nil || result.Validate() != nil {
		if err != nil {
			return User{}, err
		}
		return User{}, ErrStorage
	}
	return result, nil
}

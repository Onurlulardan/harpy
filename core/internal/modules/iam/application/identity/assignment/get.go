package assignment

import (
	"context"
)

func (commands *Commands) GetAssignment(
	ctx context.Context, scope TenantScope, query GetAssignmentQuery,
) (Assignment, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return Assignment{}, ErrInvalidQuery
	}
	result, err := commands.store.GetAssignment(ctx, scope, query)
	if err != nil {
		return Assignment{}, err
	}
	if result.Validate() != nil {
		return Assignment{}, ErrStorage
	}
	return result, nil
}

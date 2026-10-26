package assignment

import (
	"context"
)

func (commands *Commands) ListAssignments(
	ctx context.Context, scope TenantScope, query ListAssignmentsQuery,
) ([]Assignment, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return nil, ErrInvalidQuery
	}
	result, err := commands.store.ListAssignments(ctx, scope, query)
	if err != nil || len(result) > query.Limit {
		return nil, errOrStorage(err)
	}
	for _, assignment := range result {
		if assignment.Validate() != nil {
			return nil, ErrStorage
		}
	}
	return result, nil
}

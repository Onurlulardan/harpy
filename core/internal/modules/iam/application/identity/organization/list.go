package organization

import (
	"context"
)

func (commands *Commands) ListOrganizations(
	ctx context.Context,
	scope TenantScope,
	query ListOrganizationsQuery,
) ([]Organization, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return nil, ErrInvalidQuery
	}
	result, err := commands.store.ListOrganizations(ctx, scope, query)
	if err != nil || len(result) > query.Limit {
		return nil, errOrStorage(err)
	}
	for _, organization := range result {
		if organization.Validate() != nil {
			return nil, ErrStorage
		}
	}
	return result, nil
}

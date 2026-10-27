package organization

import (
	"context"
)

func (commands *Commands) GetOrganization(
	ctx context.Context,
	scope TenantScope,
	query GetOrganizationQuery,
) (Organization, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return Organization{}, ErrInvalidQuery
	}
	result, err := commands.store.GetOrganization(ctx, scope, query)
	if err != nil || result.Validate() != nil {
		if err != nil {
			return Organization{}, err
		}
		return Organization{}, ErrStorage
	}
	return result, nil
}

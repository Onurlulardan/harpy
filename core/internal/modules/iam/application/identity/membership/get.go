package membership

import (
	"context"
)

func (commands *Commands) GetMembership(
	ctx context.Context, scope TenantScope, query GetMembershipQuery,
) (Membership, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return Membership{}, ErrInvalidQuery
	}
	result, err := commands.store.GetMembership(ctx, scope, query)
	if err != nil {
		return Membership{}, err
	}
	if result.Validate() != nil {
		return Membership{}, ErrStorage
	}
	return result, nil
}

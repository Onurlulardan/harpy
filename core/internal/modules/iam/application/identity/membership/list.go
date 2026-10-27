package membership

import (
	"context"
)

func (commands *Commands) ListMemberships(
	ctx context.Context, scope TenantScope, query ListMembershipsQuery,
) ([]Membership, error) {
	if commands == nil || query.Validate() != nil ||
		!queryScopeUsable(ctx, scope, query.Evidence) {
		return nil, ErrInvalidQuery
	}
	result, err := commands.store.ListMemberships(ctx, scope, query)
	if err != nil || len(result) > query.Limit {
		return nil, errOrStorage(err)
	}
	for _, membership := range result {
		if membership.Validate() != nil {
			return nil, ErrStorage
		}
	}
	return result, nil
}

package lifecycleevent

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) Tenant(
	ctx context.Context,
	scope tenantScope,
	change lifecycle.Change,
) (int64, error) {
	if change.Validate() != nil || !tenantScopeUsable(ctx, scope) {
		return 0, lifecycle.ErrEffects
	}
	bound, err := store.bindTenant(scope)
	return store.sequence(ctx, bound, change, err)
}

func (store *store) Credential(
	ctx context.Context,
	scope credentialScope,
	change lifecycle.Change,
) (int64, error) {
	if change.Validate() != nil || !credentialScopeUsable(ctx, scope) {
		return 0, lifecycle.ErrEffects
	}
	bound, err := store.bindCredential(scope)
	return store.sequence(ctx, bound, change, err)
}

func (store *store) sequence(
	ctx context.Context,
	bound *boundStore,
	change lifecycle.Change,
	bindErr error,
) (int64, error) {
	if bindErr != nil || bound == nil {
		return 0, lifecycle.ErrEffects
	}
	var sequence int64
	err := bound.database.QueryRow(
		ctx,
		sequenceSQL,
		change.OperationID,
		change.CommandKey,
		change.AggregateKind,
		change.AggregateID,
	).Scan(&sequence)
	if err != nil || sequence < 1 {
		return 0, errors.Join(lifecycle.ErrEffects, err)
	}
	return sequence, nil
}

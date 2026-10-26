package demomaterialization

import (
	"context"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Target(
	ctx context.Context,
	scope platformtx.TenantScope,
	component contract.Component,
) (contract.Component, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validComponent(component, false) ||
		!platformtx.ContextOwnsScope(ctx, scope) || platformtx.NetworkAllowed(ctx) ||
		scope.Mode() != platformtx.Command ||
		platformtx.RequireOwner(scope, owner) != nil {
		return contract.Component{}, ErrInvalid
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || authority.Generation() < 1 ||
		component.Generation > 0 && component.Generation != authority.Generation() {
		return contract.Component{}, ErrUnavailable
	}
	component.Generation = authority.Generation()
	return component, nil
}

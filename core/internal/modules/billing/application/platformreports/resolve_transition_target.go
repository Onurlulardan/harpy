package platformreport

import (
	"context"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type Resolver struct{ binder TransitionTargetBinder }

func NewResolver(binder TransitionTargetBinder) (*Resolver, error) {
	if binder == nil {
		return nil, ErrConfiguration
	}
	return &Resolver{binder: binder}, nil
}

func (resolver *Resolver) ResolveTransitionTarget(
	ctx context.Context,
	scope platformtx.GlobalScope,
	authority platformtx.OwnerAuthority,
	subscriptionID uuid.UUID,
) (TransitionTarget, error) {
	if resolver == nil || resolver.binder == nil || ctx == nil || ctx.Err() != nil ||
		subscriptionID == uuid.Nil {
		return TransitionTarget{}, ErrConfiguration
	}
	store, err := resolver.binder.BindTransitionTarget(scope, authority)
	if err != nil {
		return TransitionTarget{}, fmt.Errorf("bind platform subscription target: %w", err)
	}
	if store == nil {
		return TransitionTarget{}, ErrConfiguration
	}
	target, err := store.ResolveTransitionTarget(ctx, subscriptionID)
	if err != nil {
		return TransitionTarget{}, fmt.Errorf("resolve platform subscription target: %w", err)
	}
	if !target.Valid() {
		return TransitionTarget{}, ErrTarget
	}
	return target, nil
}

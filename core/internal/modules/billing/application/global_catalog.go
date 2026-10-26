package application

import (
	"context"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type globalCoordinator interface {
	InGlobal(
		context.Context,
		platformtx.GlobalScopeRequest,
		func(context.Context, platformtx.GlobalScope) error,
	) error
}

type GlobalCatalog struct{ coordinator globalCoordinator }

type GlobalSnapshotAction func(context.Context, platformtx.GlobalScope) error

type GlobalCommandAction func(
	context.Context,
	platformtx.GlobalScope,
	platformtx.OwnerAuthority,
) error

func NewGlobalCatalog(coordinator globalCoordinator) (*GlobalCatalog, error) {
	if coordinator == nil {
		return nil, fmt.Errorf("global Billing coordinator is required")
	}
	return &GlobalCatalog{coordinator: coordinator}, nil
}

func (catalog *GlobalCatalog) PublicSnapshot(
	ctx context.Context,
	operation platformtx.PublicGlobalOperation,
	groupGeneration, billingGeneration int64,
	action GlobalSnapshotAction,
) error {
	if catalog == nil || catalog.coordinator == nil || action == nil {
		return platformtx.ErrInvalidGlobalScopeRequest
	}
	capability, err := platformtx.NewPublicGlobalSnapshotCapabilityIssuer().Issue(
		operation, groupGeneration, billingGeneration,
	)
	if err != nil {
		return platformtx.ErrInvalidGlobalScopeRequest
	}
	request, err := platformtx.NewPublicGlobalSnapshotScopeRequest(capability)
	if err != nil {
		return err
	}
	return catalog.coordinator.InGlobal(ctx, request, action)
}

func (catalog *GlobalCatalog) Snapshot(
	ctx context.Context,
	capability platformtx.GlobalSnapshotCapability,
	action GlobalSnapshotAction,
) error {
	if catalog == nil || catalog.coordinator == nil || action == nil {
		return platformtx.ErrInvalidGlobalScopeRequest
	}
	request, err := platformtx.NewGlobalSnapshotScopeRequest(capability)
	if err != nil {
		return err
	}
	return catalog.coordinator.InGlobal(ctx, request, action)
}

func (catalog *GlobalCatalog) Command(
	ctx context.Context,
	capability platformtx.GlobalCommandCapability,
	action GlobalCommandAction,
) error {
	if catalog == nil || catalog.coordinator == nil || action == nil {
		return platformtx.ErrInvalidGlobalScopeRequest
	}
	request, err := platformtx.NewGlobalCommandScopeRequest(capability)
	if err != nil {
		return err
	}
	return catalog.coordinator.InGlobal(
		ctx, request,
		func(scoped context.Context, scope platformtx.GlobalScope) error {
			authority, authorityErr := platformtx.MintGlobalOwnerAuthority(scope, "billing")
			if authorityErr != nil || !authority.OwnsGlobalScope(scope) {
				return platformtx.ErrScopeOwner
			}
			return action(scoped, scope, authority)
		},
	)
}

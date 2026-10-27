package catalogcomponent

import (
	"context"
	"fmt"
	"reflect"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type PersistentOwner struct {
	name   string
	binder StoreBinder
}

func NewPersistentOwner(name string, binder StoreBinder) (*PersistentOwner, error) {
	if !ownerPattern.MatchString(name) || nilPort(binder) {
		return nil, fmt.Errorf("catalog component owner dependencies are invalid")
	}
	return &PersistentOwner{name: name, binder: binder}, nil
}

func (owner *PersistentOwner) Readiness(ctx context.Context, scope platformtx.TenantScope,
	component contract.Component,
) (contract.Readiness, error) {
	store, err := owner.bind(ctx, scope, component)
	if err != nil {
		return contract.Readiness{}, err
	}
	value, err := store.Read(ctx, component)
	if err != nil || value.Component != component || value.Generation < 1 || !value.WriterV2 {
		return contract.Readiness{}, fmt.Errorf("component owner readiness is unavailable")
	}
	return value, nil
}

func (owner *PersistentOwner) Stage(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "stage")
}

func (owner *PersistentOwner) Cleanup(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "cleanup")
}

func (owner *PersistentOwner) Activate(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "activate")
}

func (owner *PersistentOwner) Disable(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "disable")
}

func (owner *PersistentOwner) Enable(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "enable")
}

func (owner *PersistentOwner) Retire(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "retire")
}

func (owner *PersistentOwner) Restore(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "restore")
}

func (owner *PersistentOwner) Purge(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.Receipt, error) {
	return owner.apply(ctx, scope, command, "purge")
}

func (owner *PersistentOwner) PurgeBlockers(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command,
) (contract.BlockerReceipt, error) {
	store, err := owner.bind(ctx, scope, command.Component)
	if err != nil || command.Action != "purge" || command.ReferenceEpoch < 1 {
		return contract.BlockerReceipt{}, fmt.Errorf("component owner purge request is invalid")
	}
	return store.Blockers(ctx, command)
}

func (owner *PersistentOwner) AllowsMenuTarget(ctx context.Context, scope platformtx.TenantScope,
	kind, target string,
) (bool, error) {
	component := contract.Component{ID: target, Owner: owner.name}
	store, err := owner.bind(ctx, scope, component)
	if err != nil {
		return false, err
	}
	return store.AllowsMenuTarget(ctx, kind, target)
}

func (owner *PersistentOwner) apply(ctx context.Context, scope platformtx.TenantScope,
	command contract.Command, action string,
) (contract.Receipt, error) {
	store, err := owner.bind(ctx, scope, command.Component)
	if err != nil || command.OperationID == "" || command.IdempotencyKey == "" ||
		command.Action != action {
		return contract.Receipt{}, fmt.Errorf("component owner command is invalid")
	}
	value, err := store.Apply(ctx, command)
	if err != nil {
		return contract.Receipt{}, fmt.Errorf("component owner apply failed: %w", err)
	}
	switch {
	case value.OperationID != command.OperationID:
		return contract.Receipt{}, fmt.Errorf("component owner receipt operation is invalid")
	case value.Component != command.Component:
		return contract.Receipt{}, fmt.Errorf("component owner receipt component is invalid")
	case value.Action != action:
		return contract.Receipt{}, fmt.Errorf("component owner receipt action is invalid")
	case value.ID == "":
		return contract.Receipt{}, fmt.Errorf("component owner receipt identity is invalid")
	case value.CompletedAt.IsZero():
		return contract.Receipt{}, fmt.Errorf("component owner receipt completion is invalid")
	}
	return value, nil
}

func (owner *PersistentOwner) bind(ctx context.Context, scope platformtx.TenantScope,
	component contract.Component,
) (Store, error) {
	if owner == nil || nilPort(owner.binder) || ctx == nil || ctx.Err() != nil ||
		component.ID == "" || component.Owner != owner.name ||
		!platformtx.ContextOwnsScope(ctx, scope) || platformtx.NetworkAllowed(ctx) ||
		platformtx.RequireOwner(scope, owner.name) != nil {
		return nil, fmt.Errorf("component owner scope is unavailable")
	}
	store, err := owner.binder.Bind(ctx, scope)
	if err != nil || nilPort(store) {
		return nil, fmt.Errorf("component owner store is unavailable")
	}
	return store, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return (seen.Kind() == reflect.Pointer || seen.Kind() == reflect.Interface) && seen.IsNil()
}

package componentowner

import (
	"context"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type LifecycleAdapter struct {
	Owner
	owner string
}

func NewLifecycleAdapter(owner string, base Owner) (*LifecycleAdapter, error) {
	if owner == "" || base == nil {
		return nil, ErrInvalid
	}
	return &LifecycleAdapter{Owner: base, owner: owner}, nil
}

func (adapter *LifecycleAdapter) RetireComponent(
	ctx context.Context,
	scope platformtx.TenantScope,
	command LifecycleCommand,
) (OwnerResult, error) {
	return adapter.applyLifecycle(ctx, scope, command, "retire")
}

func (adapter *LifecycleAdapter) RestoreComponent(
	ctx context.Context,
	scope platformtx.TenantScope,
	command LifecycleCommand,
) (OwnerResult, error) {
	return adapter.applyLifecycle(ctx, scope, command, "restore")
}

func (adapter *LifecycleAdapter) applyLifecycle(
	ctx context.Context,
	scope platformtx.TenantScope,
	command LifecycleCommand,
	action string,
) (OwnerResult, error) {
	if adapter == nil || adapter.Owner == nil || command.Action != action ||
		command.Component.Owner != adapter.owner || !command.Valid() {
		return OwnerResult{}, ErrInvalid
	}
	legacy := Command{
		OperationID: command.OperationID, Component: command.Component,
		Action: action, IdempotencyKey: command.IdempotencyKey,
	}
	var receipt Receipt
	var err error
	if action == "retire" {
		receipt, err = adapter.Owner.Retire(ctx, scope, legacy)
	} else {
		receipt, err = adapter.Owner.Restore(ctx, scope, legacy)
	}
	if err != nil || !legacyReceiptMatches(receipt, legacy) {
		return OwnerResult{}, errorOr(err, ErrUnavailable)
	}
	state := catalogcomponentv2.StateRetired
	if action == "restore" {
		state = catalogcomponentv2.StateRestored
	}
	result := OwnerResult{
		ReceiptID: receipt.ID, TenantID: command.TenantID,
		PackageID: command.PackageID, Component: command.Component,
		OperationID: command.OperationID, ReferenceEpoch: command.ReferenceEpoch,
		State: state, Ready: true, CompletedAt: receipt.CompletedAt,
	}
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	if !result.ValidForLifecycle(command) {
		return OwnerResult{}, ErrUnavailable
	}
	return result, nil
}

func (adapter *LifecycleAdapter) BeginComponentPurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PurgeCommand,
) (OwnerResult, error) {
	result, err := adapter.preflightPurge(ctx, scope, command)
	if err != nil {
		return OwnerResult{}, err
	}
	if result.State == catalogcomponentv2.StateReady {
		result.State = catalogcomponentv2.StatePurgePending
		result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	}
	if !result.ValidBeginFor(command) {
		return OwnerResult{}, ErrUnavailable
	}
	return result, nil
}

func (adapter *LifecycleAdapter) PreflightComponentPurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PurgeCommand,
) (OwnerResult, error) {
	result, err := adapter.preflightPurge(ctx, scope, command)
	if err != nil {
		return OwnerResult{}, err
	}
	if !result.ValidPreflightFor(command) {
		return OwnerResult{}, ErrUnavailable
	}
	return result, nil
}

func (adapter *LifecycleAdapter) preflightPurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PurgeCommand,
) (OwnerResult, error) {
	if !adapter.validPurge(command) {
		return OwnerResult{}, ErrInvalid
	}
	legacy := adapter.purgeCommand(command, "blockers")
	receipt, err := adapter.Owner.PurgeBlockers(ctx, scope, legacy)
	if err != nil || receipt.OperationID != legacy.OperationID ||
		receipt.ComponentID != legacy.Component.ID ||
		receipt.ReferenceEpoch != legacy.ReferenceEpoch {
		return OwnerResult{}, errorOr(err, ErrUnavailable)
	}
	result := OwnerResult{
		ReceiptID: receipt.ID, TenantID: command.Authority.TenantID,
		PackageID: command.Authority.PackageID, Component: command.Component,
		OperationID:     command.Authority.OperationID,
		ReferenceEpoch:  command.Authority.ReferenceEpoch,
		AuthoritySHA256: command.Authority.AuthoritySHA256,
	}
	if len(receipt.Blockers) != 0 {
		result.State, result.Blockers = catalogcomponentv2.StateBlocked, receipt.Blockers
	} else {
		result.State, result.Ready = catalogcomponentv2.StateReady, true
		result.CompletedAt = receipt.CompletedAt
	}
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	return result, nil
}

func (adapter *LifecycleAdapter) FinalizeComponentPurge(
	ctx context.Context,
	scope platformtx.TenantScope,
	command PurgeCommand,
) (OwnerResult, error) {
	if !adapter.validPurge(command) {
		return OwnerResult{}, ErrInvalid
	}
	legacy := adapter.purgeCommand(command, "purge")
	receipt, err := adapter.Owner.Purge(ctx, scope, legacy)
	if err != nil || !legacyReceiptMatches(receipt, legacy) {
		return OwnerResult{}, errorOr(err, ErrUnavailable)
	}
	result := OwnerResult{
		ReceiptID: receipt.ID, TenantID: command.Authority.TenantID,
		PackageID: command.Authority.PackageID, Component: command.Component,
		OperationID:     command.Authority.OperationID,
		ReferenceEpoch:  command.Authority.ReferenceEpoch,
		State:           catalogcomponentv2.StatePurged,
		AuthoritySHA256: command.Authority.AuthoritySHA256,
		Ready:           true, CompletedAt: receipt.CompletedAt,
	}
	result.ResultSHA256 = catalogcomponentv2.ResultSHA256(result)
	if !result.ValidFinalizeFor(command) {
		return OwnerResult{}, ErrUnavailable
	}
	return result, nil
}

func (adapter *LifecycleAdapter) validPurge(command PurgeCommand) bool {
	return adapter != nil && adapter.Owner != nil && command.Valid() &&
		command.Component.Owner == adapter.owner
}

func (adapter *LifecycleAdapter) purgeCommand(
	command PurgeCommand,
	action string,
) Command {
	return Command{
		OperationID: command.Authority.OperationID, Component: command.Component,
		Action: action, ReferenceEpoch: command.Authority.ReferenceEpoch,
		IdempotencyKey: command.IdempotencyKey,
	}
}

var _ LifecycleOwner = (*LifecycleAdapter)(nil)

package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) assignTask(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire, taskID ports.ID, idErr error,
) (ports.OperationReceipt, bool, error) {
	if idErr != nil {
		return ports.OperationReceipt{}, true, idErr
	}
	version, err := runtime.taskVersion(ctx, scope, taskID,
		wire.ExpectedVersion, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	principals := make([]ports.Principal, 0, len(wire.Principals))
	for _, item := range wire.Principals {
		id, parseErr := parseID(item.ID)
		principal := ports.Principal{Kind: ports.PrincipalKind(item.Kind), ID: id}
		if parseErr != nil || !principal.Valid() {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		principals = append(principals, principal)
	}
	receipt, err := runtime.services.tasks.ChangeAssignees(ctx, scope,
		mutation(input, version), taskID, principals)
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) transitionTask(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire, taskID ports.ID, idErr error,
) (ports.OperationReceipt, bool, error) {
	if idErr != nil {
		return ports.OperationReceipt{}, true, idErr
	}
	version, err := runtime.taskVersion(ctx, scope, taskID,
		wire.ExpectedVersion, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	meta := mutation(input, version)
	var receipt ports.OperationReceipt
	switch input.Operation {
	case "completeWorkTask":
		receipt, err = runtime.services.tasks.CompleteTask(ctx, scope, meta, taskID)
	case "cancelWorkTask":
		receipt, err = runtime.services.tasks.CancelTask(ctx, scope, meta, taskID)
	default:
		receipt, err = runtime.services.tasks.ReopenTask(ctx, scope, meta, taskID)
	}
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) taskLifecycle(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire, taskID ports.ID, archive bool, idErr error,
) (ports.OperationReceipt, bool, error) {
	if idErr != nil {
		return ports.OperationReceipt{}, true, idErr
	}
	version, err := runtime.taskVersion(ctx, scope, taskID,
		wire.ExpectedVersion, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	var receipt ports.OperationReceipt
	if archive {
		receipt, err = runtime.services.tasks.ArchiveTask(
			ctx, scope, mutation(input, version), taskID)
	} else {
		receipt, err = runtime.services.tasks.RestoreTask(
			ctx, scope, mutation(input, version), taskID)
	}
	return receipt, true, err
}

package work

import (
	"context"
	"encoding/json"

	workpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *PlatformHTTPRuntime) Write(ctx context.Context,
	input HTTPWriteInput,
) (result json.RawMessage, err error) {
	if !validHTTPAccess(input.Access, input.Operation, input.Resource, input.Compatibility) {
		return nil, ErrHTTPForbidden
	}
	wire, err := decodeWrite(input.Body)
	if err != nil {
		return nil, err
	}
	values, err := runtime.resolve(ctx, input.Access)
	if err != nil {
		return nil, err
	}
	formsScope := needsFormsTaskScope(input, wire)
	if formsScope {
		values, err = runtime.resolveForms(ctx, input.Access, values)
		if err != nil {
			return nil, err
		}
	}
	dispatch := func(operation context.Context, scope platformtx.TenantScope) error {
		receipt, dispatchErr := runtime.dispatchWrite(operation,
			workpostgres.BridgeScope(scope), input, wire)
		if dispatchErr == nil {
			result, dispatchErr = receiptDocument(receipt, input.Compatibility)
		}
		return dispatchErr
	}
	if formsScope {
		err = runtime.commandForms(ctx, input.Access, values, input.Access.Authority(), dispatch)
	} else {
		err = runtime.command(ctx, input.Access, values, input.Access.Authority(), dispatch)
	}
	return result, err
}

func needsFormsTaskScope(input HTTPWriteInput, wire writeWire) bool {
	switch input.Operation {
	case "createWorkTask", "createLegacyBoardTask":
		kind := wire.Kind
		if input.Compatibility {
			kind = wire.TaskType
		}
		return kind == string(ports.TaskFormFill) ||
			kind == string(ports.TaskFormReview) || wire.Present["form_slug"] ||
			wire.Present["form_data_id"]
	case "updateWorkTask", "updateLegacyBoardTask":
		return wire.Present["task_type"] || wire.Present["task_kind"] ||
			wire.Present["form_slug"] || wire.Present["form_data_id"]
	default:
		return false
	}
}

func (runtime *PlatformHTTPRuntime) dispatchWrite(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, error) {
	if receipt, matched, err := runtime.writeBoardTask(ctx, scope, input, wire); matched {
		return receipt, err
	}
	if receipt, matched, err := runtime.writeCollaboration(ctx, scope, input, wire); matched {
		return receipt, err
	}
	return ports.OperationReceipt{}, ports.ErrInvalid
}

package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/columns"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) boardLifecycle(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire, boardID ports.ID, archive bool, idErr error,
) (ports.OperationReceipt, bool, error) {
	if idErr != nil {
		return ports.OperationReceipt{}, true, idErr
	}
	version, err := runtime.boardVersion(ctx, scope, boardID,
		wire.ExpectedVersion, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	var receipt ports.OperationReceipt
	if archive {
		receipt, err = runtime.services.boards.ArchiveBoard(
			ctx, scope, mutation(input, version), boardID)
	} else {
		receipt, err = runtime.services.boards.RestoreBoard(
			ctx, scope, mutation(input, version), boardID)
	}
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) updateColumn(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	columnID, err := routeID(input.IDs, "column_id")
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	version, name, color, status, err := runtime.columnValues(ctx, scope, columnID,
		wire.ExpectedVersion, wire.Name, wire.Color, wire.Status,
		wire.LinkedStatus, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	receipt, err := runtime.services.columns.UpdateColumn(ctx, scope,
		mutation(input, version), columns.UpdateInput{
			ColumnID: columnID, Name: name, Color: color, Status: status})
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) columnValues(ctx context.Context, scope ports.Scope,
	columnID ports.ID, version int64, name, color, status, linked string, compatibility bool,
) (int64, string, string, ports.TaskStatus, error) {
	wanted := ports.TaskStatus(status)
	if !wanted.Valid() && linked != "" {
		wanted = legacyColumnStatus(linked)
	}
	if !compatibility || version > 0 {
		if color == "" {
			color = "#64748b"
		}
		return version, name, color, wanted, nil
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, "", "", "", err
	}
	column, err := stores.Boards().LoadColumnForUpdate(ctx, columnID)
	if name == "" {
		name = column.Name
	}
	if color == "" {
		color = column.Color
	}
	if !wanted.Valid() {
		wanted = column.Status
	}
	return column.RowVersion, name, color, wanted, err
}

func (runtime *PlatformHTTPRuntime) reorderColumns(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire, boardID ports.ID, idErr error,
) (ports.OperationReceipt, bool, error) {
	if idErr != nil || len(wire.ColumnIDs) == 0 {
		return ports.OperationReceipt{}, true, ports.ErrInvalid
	}
	ids := make([]ports.ID, 0, len(wire.ColumnIDs))
	for _, raw := range wire.ColumnIDs {
		id, err := parseID(raw)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		ids = append(ids, id)
	}
	version, err := runtime.boardVersion(ctx, scope, boardID,
		wire.ExpectedVersion, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	receipt, err := runtime.services.columns.ReorderColumns(ctx, scope,
		mutation(input, version), columns.ReorderInput{BoardID: boardID, ColumnIDs: ids})
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) deleteColumn(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire, boardID ports.ID, boardErr error,
) (ports.OperationReceipt, bool, error) {
	columnID, err := routeID(input.IDs, "column_id")
	if boardErr != nil || err != nil {
		return ports.OperationReceipt{}, true, ports.ErrInvalid
	}
	version, err := runtime.columnVersion(ctx, scope, columnID,
		wire.ExpectedVersion, input.Compatibility)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	receipt, err := runtime.services.columns.DeleteColumn(ctx, scope,
		mutation(input, version), columns.DeleteInput{BoardID: boardID, ColumnID: columnID})
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) columnVersion(ctx context.Context, scope ports.Scope,
	columnID ports.ID, requested int64, compatibility bool,
) (int64, error) {
	if requested > 0 || !compatibility {
		return requested, nil
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, err
	}
	column, err := stores.Boards().LoadColumnForUpdate(ctx, columnID)
	return column.RowVersion, err
}

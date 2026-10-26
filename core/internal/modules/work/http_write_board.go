package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/boards"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/columns"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeBoard(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	boardID, err := routeID(input.IDs, "board_id")
	switch input.Operation {
	case "createWorkBoard", "createLegacyWorkBoard":
		initial := wire.InitialColumn
		if initial == "" {
			initial = "Backlog"
		}
		receipt, writeErr := runtime.services.boards.CreateBoard(ctx, scope,
			mutation(input, 0), boards.CreateBoardInput{Name: wire.Name,
				Description: wire.Description, InitialColumn: initial})
		return receipt, true, writeErr
	case "updateWorkBoard", "updateLegacyWorkBoard":
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		version, name, description, readErr := runtime.boardUpdateValues(
			ctx, scope, boardID, input.Compatibility, wire)
		if readErr != nil {
			return ports.OperationReceipt{}, true, readErr
		}
		receipt, writeErr := runtime.services.boards.UpdateBoard(ctx, scope,
			mutation(input, version), boardID,
			boards.UpdateBoardInput{Name: name, Description: description})
		return receipt, true, writeErr
	case "archiveWorkBoard", "archiveLegacyWorkBoard":
		return runtime.boardLifecycle(ctx, scope, input, wire, boardID, true, err)
	case "restoreWorkBoard":
		return runtime.boardLifecycle(ctx, scope, input, wire, boardID, false, err)
	case "createWorkColumn", "createLegacyBoardColumn":
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		version, readErr := runtime.boardVersion(ctx, scope, boardID,
			wire.ExpectedVersion, input.Compatibility)
		status := ports.TaskStatus(wire.Status)
		if !status.Valid() {
			status = legacyColumnStatus(wire.LinkedStatus)
		}
		color := wire.Color
		if color == "" {
			color = "#64748b"
		}
		before, beforeErr := optionalID(wire.BeforeID)
		after, afterErr := optionalID(wire.AfterID)
		if readErr != nil || beforeErr != nil || afterErr != nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		receipt, writeErr := runtime.services.columns.CreateColumn(ctx, scope,
			mutation(input, version), columns.CreateColumnInput{BoardID: boardID,
				Name: wire.Name, Color: color, Status: status, Before: before, After: after})
		return receipt, true, writeErr
	case "updateWorkColumn", "updateLegacyBoardColumn":
		return runtime.updateColumn(ctx, scope, input, wire)
	case "reorderLegacyBoardColumns":
		return runtime.reorderColumns(ctx, scope, input, wire, boardID, err)
	case "deleteLegacyBoardColumn":
		return runtime.deleteColumn(ctx, scope, input, wire, boardID, err)
	}
	return ports.OperationReceipt{}, false, nil
}

func legacyColumnStatus(value string) ports.TaskStatus {
	switch value {
	case "IN_PROGRESS":
		return ports.StatusInProgress
	case "IN_REVIEW":
		return ports.StatusInReview
	case "DONE":
		return ports.StatusDone
	default:
		return ports.StatusOpen
	}
}

package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeCommentLabel(ctx context.Context,
	scope ports.Scope, input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	if receipt, matched, err := runtime.writeComment(ctx, scope, input, wire); matched {
		return receipt, true, err
	}
	switch input.Operation {
	case "createWorkLabel":
		boardID, err := parseID(wire.BoardID)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		receipt, err := runtime.services.labels.CreateLabel(ctx, scope,
			mutation(input, 0), boardID, wire.Name, wire.Color)
		return receipt, true, err
	case "updateWorkLabel":
		labelID, labelErr := routeID(input.IDs, "label_id")
		boardID, boardErr := parseID(wire.BoardID)
		if labelErr != nil || boardErr != nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		receipt, err := runtime.services.labels.UpdateLabel(ctx, scope,
			mutation(input, wire.ExpectedVersion), ports.Label{ID: labelID,
				BoardID: boardID, Name: wire.Name, Color: wire.Color})
		return receipt, true, err
	case "setWorkTaskLabel":
		taskID, taskErr := routeID(input.IDs, "task_id")
		labelID, labelErr := routeID(input.IDs, "label_id")
		if taskErr != nil || labelErr != nil || wire.Linked == nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		receipt, err := runtime.services.labels.SetTaskLabel(ctx, scope,
			mutation(input, wire.ExpectedVersion), taskID, labelID, *wire.Linked)
		return receipt, true, err
	}
	return ports.OperationReceipt{}, false, nil
}

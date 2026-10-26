package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/checklists"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeChecklist(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	switch input.Operation {
	case "createWorkChecklistItem":
		taskID, err := routeID(input.IDs, "task_id")
		before, beforeErr := optionalID(wire.BeforeID)
		after, afterErr := optionalID(wire.AfterID)
		if err != nil || beforeErr != nil || afterErr != nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		receipt, err := runtime.services.checklists.CreateChecklistItem(ctx, scope,
			mutation(input, 0), checklists.CreateChecklistInput{TaskID: taskID,
				TaskVersion: wire.TaskVersion, Text: wire.Text,
				Before: before, After: after})
		return receipt, true, err
	case "updateWorkChecklistItem":
		itemID, itemErr := routeID(input.IDs, "checklist_id")
		taskID, taskErr := parseID(wire.TaskID)
		if itemErr != nil || taskErr != nil || wire.Completed == nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		receipt, err := runtime.services.checklists.UpdateChecklistItem(ctx, scope,
			mutation(input, wire.ExpectedVersion), checklists.UpdateChecklistInput{
				Item: ports.ChecklistItem{ID: itemID, TaskID: taskID,
					Text: wire.Text, Completed: *wire.Completed, Position: wire.Position},
				TaskVersion: wire.TaskVersion})
		return receipt, true, err
	}
	return ports.OperationReceipt{}, false, nil
}

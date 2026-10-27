package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeBoardTask(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	if receipt, matched, err := runtime.writeBoard(ctx, scope, input, wire); matched {
		return receipt, true, err
	}
	taskID, taskErr := routeID(input.IDs, "task_id")
	switch input.Operation {
	case "createWorkTask", "createLegacyBoardTask":
		return runtime.createTask(ctx, scope, input, wire)
	case "updateWorkTask", "updateLegacyBoardTask":
		if taskErr != nil {
			return ports.OperationReceipt{}, true, taskErr
		}
		values, err := runtime.taskUpdateValues(
			ctx, scope, taskID, input.Compatibility, wire)
		if err != nil {
			return ports.OperationReceipt{}, true, err
		}
		var intentHash [32]byte
		if values.CompatibilityImplicitVersion {
			intentHash, err = legacyTaskUpdateIntentHash(taskID, input.Body)
			if err != nil {
				return ports.OperationReceipt{}, true, err
			}
		}
		receipt, err := runtime.services.tasks.UpdateTask(ctx, scope,
			mutation(input, values.Version), tasks.UpdateInput{TaskID: taskID,
				Title: values.Title, Description: values.Description,
				Priority: values.Priority, StartsAt: values.StartsAt,
				DueAt: values.DueAt, Relations: values.Relations,
				ReplaceResource: values.ReplaceResource, Kind: values.Kind,
				Resource:                     values.Resource,
				ReplaceLegacyMetadata:        values.ReplaceLegacyMetadata,
				LegacyMetadata:               values.LegacyMetadata,
				CompatibilityImplicitVersion: values.CompatibilityImplicitVersion,
				CompatibilityIntentHash:      intentHash})
		return receipt, true, err
	case "moveWorkTask", "moveLegacyBoardTask":
		if taskErr != nil {
			return ports.OperationReceipt{}, true, taskErr
		}
		columnID, err := parseID(wire.ColumnID)
		version, versionErr := runtime.taskVersion(ctx, scope, taskID,
			wire.ExpectedVersion, input.Compatibility)
		if err != nil || versionErr != nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		if input.Compatibility {
			order := wire.Order
			receipt, moveErr := runtime.services.tasks.MoveTask(ctx, scope,
				mutation(input, version), tasks.MoveTaskInput{TaskID: taskID,
					ColumnID: columnID, CompatibilityOrder: &order})
			return receipt, true, moveErr
		}
		before, beforeErr := optionalID(wire.BeforeID)
		after, afterErr := optionalID(wire.AfterID)
		if beforeErr != nil || afterErr != nil {
			return ports.OperationReceipt{}, true, ports.ErrInvalid
		}
		receipt, err := runtime.services.tasks.MoveTask(ctx, scope, mutation(input, version),
			tasks.MoveTaskInput{TaskID: taskID, ColumnID: columnID,
				Before: before, After: after})
		return receipt, true, err
	case "assignWorkTask":
		return runtime.assignTask(ctx, scope, input, wire, taskID, taskErr)
	case "completeWorkTask", "cancelWorkTask", "reopenWorkTask":
		return runtime.transitionTask(ctx, scope, input, wire, taskID, taskErr)
	case "archiveWorkTask", "archiveLegacyBoardTask":
		return runtime.taskLifecycle(ctx, scope, input, wire, taskID, true, taskErr)
	case "restoreWorkTask":
		return runtime.taskLifecycle(ctx, scope, input, wire, taskID, false, taskErr)
	}
	return ports.OperationReceipt{}, false, nil
}

func (runtime *PlatformHTTPRuntime) createTask(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	boardID, err := routeOrBodyID(input.IDs, "board_id", wire.BoardID)
	columnID, columnErr := parseID(wire.ColumnID)
	before, beforeErr := optionalID(wire.BeforeID)
	after, afterErr := optionalID(wire.AfterID)
	if err != nil || columnErr != nil || beforeErr != nil || afterErr != nil {
		return ports.OperationReceipt{}, true, ports.ErrInvalid
	}
	kind, kindErr := taskKind(wire, input.Compatibility)
	priority, priorityErr := taskPriority(wire.Priority, ports.PriorityMedium)
	starts, due, dateErr := taskCreateTimes(wire, input.Compatibility)
	relations, relationErr := taskRelations(wire, input.Compatibility, true)
	for _, valueErr := range []error{kindErr, priorityErr, dateErr, relationErr} {
		if valueErr != nil {
			return ports.OperationReceipt{}, true, valueErr
		}
	}
	resource, resourceErr := runtime.createTaskResource(
		ctx, scope, input.Compatibility, kind, wire)
	legacy, legacyErr := createLegacyTaskMetadata(wire, input.Compatibility)
	if resourceErr != nil || legacyErr != nil {
		if resourceErr == nil {
			resourceErr = legacyErr
		}
		return ports.OperationReceipt{}, true, resourceErr
	}
	receipt, err := runtime.services.tasks.CreateTask(ctx, scope, mutation(input, 0),
		tasks.CreateTaskInput{BoardID: boardID, ColumnID: columnID, Kind: kind,
			Priority: priority, Title: wire.Title, Description: wire.Description,
			StartsAt: starts, DueAt: due, Before: before, After: after,
			Assignees: relations.Assignees, Labels: relations.Labels,
			Checklist: relations.Checklist, Resource: resource,
			LegacyMetadata: legacy})
	return receipt, true, err
}

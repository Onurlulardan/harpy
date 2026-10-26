package work

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func routeID(ids map[string]string, key string) (ports.ID, error) {
	return parseID(ids[key])
}

func routeOrBodyID(ids map[string]string, key, body string) (ports.ID, error) {
	if value := ids[key]; value != "" {
		return parseID(value)
	}
	return parseID(body)
}

func (runtime *PlatformHTTPRuntime) stores(ctx context.Context,
	scope ports.Scope,
) (ports.Stores, error) {
	return runtime.unit.Bind(ctx, scope)
}

func (runtime *PlatformHTTPRuntime) boardVersion(ctx context.Context, scope ports.Scope,
	boardID ports.ID, requested int64, compatibility bool,
) (int64, error) {
	if requested > 0 || !compatibility {
		return requested, nil
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, err
	}
	board, err := stores.Boards().LoadBoardForUpdate(ctx, boardID)
	return board.RowVersion, err
}

func (runtime *PlatformHTTPRuntime) boardUpdateValues(ctx context.Context, scope ports.Scope,
	boardID ports.ID, compatibility bool, wire writeWire,
) (int64, string, string, error) {
	if !compatibility || wire.ExpectedVersion > 0 {
		return wire.ExpectedVersion, wire.Name, wire.Description, nil
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, "", "", err
	}
	board, err := stores.Boards().LoadBoardForUpdate(ctx, boardID)
	if wire.Name == "" {
		wire.Name = board.Name
	}
	if wire.Description == "" {
		wire.Description = board.Description
	}
	return board.RowVersion, wire.Name, wire.Description, err
}

func (runtime *PlatformHTTPRuntime) taskVersion(ctx context.Context, scope ports.Scope,
	taskID ports.ID, requested int64, compatibility bool,
) (int64, error) {
	if requested > 0 || !compatibility {
		return requested, nil
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return 0, err
	}
	task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
	return task.RowVersion, err
}

type taskUpdateSnapshot struct {
	Version                      int64
	Title, Description           string
	Priority                     ports.TaskPriority
	StartsAt, DueAt              *time.Time
	Relations                    tasks.RelationInput
	ReplaceResource              bool
	Kind                         ports.TaskKind
	Resource                     *tasks.ResourceLinkInput
	ReplaceLegacyMetadata        bool
	LegacyMetadata               domain.LegacyTaskMetadata
	CompatibilityImplicitVersion bool
}

func (runtime *PlatformHTTPRuntime) taskUpdateValues(ctx context.Context, scope ports.Scope,
	taskID ports.ID, compatibility bool, wire writeWire,
) (taskUpdateSnapshot, error) {
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return taskUpdateSnapshot{}, err
	}
	task, err := stores.Tasks().LoadTaskForUpdate(ctx, taskID)
	if err != nil {
		return taskUpdateSnapshot{}, err
	}
	if compatibility && presentCompatibilityField(wire, "status") {
		status := ports.TaskStatus(wire.Status)
		if wire.Status == "TODO" {
			status = ports.StatusOpen
		}
		if !status.Valid() || status != task.Status {
			return taskUpdateSnapshot{}, ports.ErrInvalid
		}
	}
	result := taskUpdateSnapshot{Version: wire.ExpectedVersion, Title: task.Title,
		Description: task.Description, Priority: task.Priority,
		StartsAt: task.StartsAt, DueAt: task.DueAt}
	if compatibility && result.Version == 0 {
		result.Version = task.RowVersion
		result.CompatibilityImplicitVersion = true
	}
	if wire.Present["title"] && (!compatibility || !wire.Null["title"]) {
		result.Title = wire.Title
	}
	if wire.Present["description"] && (!compatibility || !wire.Null["description"]) {
		result.Description = wire.Description
	}
	if wire.Present["priority"] && (!compatibility || !wire.Null["priority"]) {
		result.Priority, err = taskPriority(wire.Priority, task.Priority)
	}
	if err == nil {
		result.StartsAt, err = patchedTaskTime(wire, compatibility,
			"starts_at", "start_date", task.StartsAt)
	}
	if err == nil {
		result.DueAt, err = patchedTaskTime(wire, compatibility,
			"due_at", "due_date", task.DueAt)
	}
	if err == nil {
		result.Relations, err = taskRelations(wire, compatibility, false)
	}
	if err == nil {
		result.ReplaceResource, result.Kind, result.Resource, err =
			runtime.updateTaskResource(ctx, scope, stores.Tasks(), task, compatibility, wire)
	}
	if err == nil {
		result.ReplaceLegacyMetadata, result.LegacyMetadata, err =
			patchLegacyTaskMetadata(ctx, stores.Tasks(), task.ID, compatibility, wire)
	}
	return result, err
}

func patchedTaskTime(wire writeWire, compatibility bool, canonical, legacy string,
	current *time.Time,
) (*time.Time, error) {
	key := canonical
	if compatibility {
		key = legacy
	}
	if !wire.Present[key] || compatibility && wire.Null[key] {
		return current, nil
	}
	if compatibility {
		if legacy == "start_date" {
			return optionalTaskTime(wire.StartDate)
		}
		return optionalTaskTime(wire.DueDate)
	}
	if canonical == "starts_at" {
		return wire.StartsAt, nil
	}
	return wire.DueAt, nil
}

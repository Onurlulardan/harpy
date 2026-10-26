package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type taskResourceLinkLoader interface {
	LoadResourceLinksForUpdate(context.Context, domain.ID) (
		[]domain.ResourceLink, error)
}

func (runtime *PlatformHTTPRuntime) createTaskResource(
	ctx context.Context,
	scope ports.Scope,
	compatibility bool,
	kind ports.TaskKind,
	wire writeWire,
) (*tasks.ResourceLinkInput, error) {
	if kind == ports.TaskCustom {
		if wire.FormSlug != "" || wire.FormDataID != "" {
			return nil, ports.ErrInvalid
		}
		return nil, nil
	}
	if !compatibility || wire.FormSlug == "" {
		return nil, ports.ErrInvalid
	}
	return runtime.taskResources.Resolve(ctx, scope, LegacyTaskResourceRequest{
		FormSlug: wire.FormSlug, FormDataID: wire.FormDataID,
	})
}

func (runtime *PlatformHTTPRuntime) updateTaskResource(
	ctx context.Context,
	scope ports.Scope,
	links taskResourceLinkLoader,
	task domain.Task,
	compatibility bool,
	wire writeWire,
) (bool, ports.TaskKind, *tasks.ResourceLinkInput, error) {
	kindField := "task_kind"
	if compatibility {
		kindField = "task_type"
	}
	present := func(field string) bool {
		return wire.Present[field] && (!compatibility || !wire.Null[field])
	}
	if !present(kindField) && !present("form_slug") && !present("form_data_id") {
		return false, "", nil, nil
	}
	desiredKind := task.Kind
	if present(kindField) {
		parsed, err := taskKind(wire, compatibility)
		if err != nil {
			return false, "", nil, err
		}
		desiredKind = parsed
	}
	if desiredKind == ports.TaskCustom {
		if wire.FormSlug != "" || wire.FormDataID != "" {
			return false, "", nil, ports.ErrInvalid
		}
		return true, desiredKind, nil, nil
	}
	if !compatibility {
		return false, "", nil, ports.ErrUnavailable
	}
	stored, err := links.LoadResourceLinksForUpdate(ctx, task.ID)
	if err != nil {
		return false, "", nil, err
	}
	current, err := tasks.ResourceInputFromLinks(task.ID, stored)
	if err != nil || task.Kind != ports.TaskCustom && current == nil {
		return false, "", nil, ports.ErrUnavailable
	}
	if present(kindField) && !present("form_slug") &&
		!present("form_data_id") && current != nil {
		return true, desiredKind, current, nil
	}
	desired := LegacyTaskResourceRequest{FormSlug: wire.FormSlug,
		FormDataID: wire.FormDataID}
	if current != nil && (!present("form_slug") || !present("form_data_id")) {
		description, describeErr := runtime.taskResources.Describe(ctx, scope, current)
		if describeErr != nil {
			return false, "", nil, describeErr
		}
		if !present("form_slug") {
			desired.FormSlug = description.FormSlug
		}
		if !present("form_data_id") {
			desired.FormDataID = description.FormDataID
		}
	}
	if desired.FormSlug == "" {
		return false, "", nil, ports.ErrInvalid
	}
	resource, err := runtime.taskResources.Resolve(ctx, scope, desired)
	if err != nil {
		return false, "", nil, err
	}
	return true, desiredKind, resource, nil
}

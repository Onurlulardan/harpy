package work

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) enrichLegacyTaskRead(
	ctx context.Context, scope ports.Scope, document json.RawMessage, boardDetail bool,
) (json.RawMessage, error) {
	parsed, err := parseLegacyTaskReadDocument(document, boardDetail)
	if err != nil || runtime == nil || runtime.reader == nil || runtime.taskResources == nil {
		return nil, ports.ErrUnavailable
	}
	resources, err := runtime.reader.ReadTaskResources(ctx, scope, parsed.taskIDs())
	if err != nil {
		return nil, err
	}
	return enrichLegacyTaskResources(ctx, scope, parsed, resources, runtime.taskResources)
}

func enrichLegacyTaskResources(
	ctx context.Context,
	scope ports.Scope,
	document *legacyTaskReadDocument,
	resources map[ports.ID]*tasks.ResourceLinkInput,
	resolver LegacyTaskResourceResolver,
) (json.RawMessage, error) {
	if ctx == nil || ctx.Err() != nil || scope == nil || document == nil || resolver == nil {
		return nil, ports.ErrUnavailable
	}
	requested := make(map[ports.ID]struct{}, len(document.tasks))
	for _, task := range document.tasks {
		requested[task.id] = struct{}{}
	}
	for taskID := range resources {
		if _, ok := requested[taskID]; !ok {
			return nil, ports.ErrUnavailable
		}
	}
	for _, task := range document.tasks {
		if err := enrichLegacyTaskResource(ctx, scope, task, resources[task.id], resolver); err != nil {
			return nil, ports.ErrUnavailable
		}
	}
	return document.marshal()
}

func enrichLegacyTaskResource(
	ctx context.Context,
	scope ports.Scope,
	task legacyTaskReadItem,
	resource *tasks.ResourceLinkInput,
	resolver LegacyTaskResourceResolver,
) error {
	if task.kind == ports.TaskCustom {
		if resource != nil {
			return ports.ErrUnavailable
		}
		return setLegacyTaskFormFields(task.values, "", "")
	}
	if !exactLegacyTaskResource(resource) {
		return ports.ErrUnavailable
	}
	description, err := resolver.Describe(ctx, scope, resource)
	if err != nil || strings.TrimSpace(description.FormSlug) == "" ||
		description.FormSlug != strings.TrimSpace(description.FormSlug) {
		return ports.ErrUnavailable
	}
	if len(resource.Targets) == 1 {
		if description.FormDataID != "" {
			return ports.ErrUnavailable
		}
	} else {
		recordID, parseErr := ports.ParseID(description.FormDataID)
		if parseErr != nil || recordID != resource.Targets[1].ID {
			return ports.ErrUnavailable
		}
	}
	return setLegacyTaskFormFields(task.values,
		description.FormSlug, description.FormDataID)
}

func exactLegacyTaskResource(resource *tasks.ResourceLinkInput) bool {
	if resource == nil || resource.Owner != ports.SourceForms ||
		len(resource.Targets) < 1 || len(resource.Targets) > 2 {
		return false
	}
	definition := resource.Targets[0]
	if definition.Kind != "FORM" || !definition.ID.Valid() ||
		!definition.VersionID.Valid() {
		return false
	}
	if len(resource.Targets) == 1 {
		return true
	}
	record := resource.Targets[1]
	return record.Kind == "FORM_RECORD" && record.ID.Valid() &&
		record.VersionID == definition.VersionID && record.ID != definition.ID
}

func setLegacyTaskFormFields(
	values map[string]json.RawMessage, slug, dataID string,
) error {
	encodedSlug, err := json.Marshal(slug)
	if err != nil {
		return ports.ErrUnavailable
	}
	values["form_slug"] = encodedSlug
	values["form_data_id"] = json.RawMessage("null")
	if dataID != "" {
		encodedID, marshalErr := json.Marshal(dataID)
		if marshalErr != nil {
			return ports.ErrUnavailable
		}
		values["form_data_id"] = encodedID
	}
	return nil
}

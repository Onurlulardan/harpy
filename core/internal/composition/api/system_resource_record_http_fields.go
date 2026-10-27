package api

import (
	"context"
	"encoding/json"
	"strings"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) mintRecordFieldValues(
	ctx context.Context,
	admission systemresources.Admission,
	resource systemresources.Resource,
	document map[string]json.RawMessage,
) (systemresources.FieldValues, error) {
	fields, err := handler.application.Fields(ctx, admission, resource)
	if err != nil {
		return systemresources.FieldValues{}, err
	}
	byName := make(map[string]systemresources.ID, len(fields))
	editable := make(map[systemresources.ID]bool, len(fields))
	for _, field := range fields {
		if !field.ID.Valid() || field.Name == "" {
			return systemresources.FieldValues{}, systemresources.ErrUnavailable
		}
		if _, duplicate := byName[field.Name]; duplicate {
			return systemresources.FieldValues{}, systemresources.ErrUnavailable
		}
		byName[field.Name] = field.ID
		editable[field.ID] = field.Editable
	}
	if err := normalizeSystemResourceDisplayName(resource, document, byName, editable); err != nil {
		return systemresources.FieldValues{}, err
	}
	return mintSystemResourceRecordFieldValues(document, byName, editable)
}

func mintSystemResourceRecordFieldValues(
	document map[string]json.RawMessage,
	byName map[string]systemresources.ID,
	editable map[systemresources.ID]bool,
) (systemresources.FieldValues, error) {
	values := make(map[systemresources.ID]json.RawMessage, len(document))
	for name, raw := range document {
		id, found := byName[name]
		if !found && systemResourceDisplayCompanion(name, document, byName) {
			continue
		}
		if !found {
			return systemresources.FieldValues{}, systemresources.ErrContract
		}
		if !editable[id] {
			continue
		}
		values[id] = append(json.RawMessage(nil), raw...)
	}
	if len(values) == 0 {
		return systemresources.FieldValues{}, systemresources.ErrContract
	}
	resultValues, err := systemresources.NewFieldValues(values)
	if err != nil {
		return systemresources.FieldValues{}, systemresources.ErrContract
	}
	return resultValues, nil
}

func systemResourceDisplayCompanion(
	name string,
	document map[string]json.RawMessage,
	known map[string]systemresources.ID,
) bool {
	base := strings.TrimSuffix(name, "_display")
	if base == name {
		return false
	}
	_, submitted := document[base]
	_, catalogued := known[base]
	return submitted && catalogued
}

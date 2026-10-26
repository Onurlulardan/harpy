package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func resolveProjections(
	catalog api.FieldCatalog,
	names []string,
	basePath string,
	includeStable bool,
	hideStable bool,
	collector *resolutionIssues,
) ([]resolved.Projection, api.FieldDefinition) {
	stableID, err := catalog.StableID()
	if err != nil {
		collector.add(basePath, restqueryv1.IssueReferenceInvalid)
		return nil, api.FieldDefinition{}
	}
	stableField, found, err := catalog.Field(stableID)
	if err != nil || !found {
		collector.add(basePath, restqueryv1.IssueReferenceInvalid)
		return nil, api.FieldDefinition{}
	}
	fields, paths := projectionFields(catalog, names, basePath, collector)
	projections := make([]resolved.Projection, 0, len(fields)+1)
	stableSelected := false
	for index, field := range fields {
		identifier, idErr := field.ID()
		stableSelected = stableSelected || idErr == nil && identifier == stableID
		projection, projectionErr := newProjection(field, paths[index], false)
		if projectionErr != nil {
			collector.add(paths[index], restqueryv1.IssueCapabilityRequired)
			continue
		}
		projections = append(projections, projection)
	}
	if !stableSelected && (includeStable || hideStable) {
		projection, projectionErr := newProjection(stableField, basePath, hideStable)
		if projectionErr != nil {
			collector.add(basePath, restqueryv1.IssueCapabilityRequired)
		} else {
			projections = append(projections, projection)
		}
	}
	return projections, stableField
}

func projectionFields(
	catalog api.FieldCatalog, names []string, basePath string,
	collector *resolutionIssues,
) ([]api.FieldDefinition, []string) {
	if len(names) != 0 {
		fields := make([]api.FieldDefinition, 0, len(names))
		paths := make([]string, 0, len(names))
		for index, name := range names {
			path := indexedPath(basePath, index)
			field, found, err := catalog.FieldByPublicName(name)
			if err != nil || !found {
				collector.add(path, restqueryv1.IssueFieldUnknown)
				continue
			}
			fields, paths = append(fields, field), append(paths, path)
		}
		return fields, paths
	}
	return defaultProjectionFields(catalog, basePath, collector)
}

func defaultProjectionFields(
	catalog api.FieldCatalog, basePath string, collector *resolutionIssues,
) ([]api.FieldDefinition, []string) {
	identifiers, err := catalog.DefaultSelect()
	if err != nil {
		collector.add(basePath, restqueryv1.IssueReferenceInvalid)
		return nil, nil
	}
	fields := make([]api.FieldDefinition, 0, len(identifiers))
	paths := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		field, found, fieldErr := catalog.Field(identifier)
		if fieldErr != nil || !found {
			collector.add(basePath, restqueryv1.IssueReferenceInvalid)
			continue
		}
		fields, paths = append(fields, field), append(paths, basePath)
	}
	return fields, paths
}

func newProjection(
	field api.FieldDefinition, path string, hidden bool,
) (resolved.Projection, error) {
	name, err := field.PublicName()
	if err != nil {
		return resolved.Projection{}, err
	}
	key, err := resolved.NewOutputKey(name)
	if err != nil {
		return resolved.Projection{}, err
	}
	return resolved.NewProjection(field, key, mustRequestPath(path), hidden)
}

func fieldUUID(field api.FieldDefinition) uuid.UUID {
	identifier, _ := field.ID()
	return identifier
}

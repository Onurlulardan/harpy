package application

import (
	"encoding/hex"
	"time"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const defaultPresentationSectionID = "00000000-0000-4000-8000-000000000001"

var defaultPresentationTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func defaultPresentation(
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
) (presentation.Document, error) {
	if !resource.Valid() || catalog.Validate() != nil {
		return presentation.Document{}, layoutapi.ErrInvalid
	}
	digest, err := catalog.CanonicalSHA256()
	if err != nil {
		return presentation.Document{}, layoutapi.ErrInvalid
	}
	definition, err := defaultPresentationDefinition(resource, catalog)
	if err != nil {
		return presentation.Document{}, err
	}
	seed := "default:" + string(resource) + ":" + hex.EncodeToString(digest[:])
	document, err := compilePresentationDefinition(
		seed, defaultPresentationTime, definition, catalog,
	)
	if err != nil {
		return presentation.Document{}, err
	}
	if _, _, err := presentation.EncodeDocument(document, catalog); err != nil {
		return presentation.Document{}, layoutapi.ErrUnavailable
	}
	return presentation.CloneDocument(document), nil
}

func defaultPresentationFields(
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
	surface layoutapi.SurfaceKind,
) ([]layoutapi.Field, error) {
	values, err := catalog.Fields()
	if err != nil {
		return nil, layoutapi.ErrInvalid
	}
	available := make(map[string]struct{}, len(values))
	custom := make([]string, 0, len(values))
	for _, field := range values {
		id, idErr := field.ID()
		if idErr != nil {
			return nil, layoutapi.ErrInvalid
		}
		available[id] = struct{}{}
		if !defaultCoreFieldID(resource, id) {
			custom = append(custom, id)
		}
	}
	ids, includeCustom, found := defaultSurfaceFieldIDs(resource, surface)
	if !found {
		return nil, layoutapi.ErrInvalid
	}
	selected := make([]string, 0, len(ids)+len(custom))
	if includeCustom {
		selected = append(selected, custom...)
	}
	for _, id := range ids {
		if _, present := available[id]; present {
			selected = append(selected, id)
		}
	}
	fields := make([]layoutapi.Field, len(selected))
	for index, id := range selected {
		fields[index] = defaultPresentationField(resource, id, int32(index))
	}
	return fields, nil
}

func defaultPresentationDefinition(
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
) (layoutapi.Definition, error) {
	profiles := systemResourceSurfaceProfiles[:5]
	surfaces := make([]layoutapi.SurfaceDefinition, len(profiles))
	for index, profile := range profiles {
		fields, err := defaultPresentationFields(resource, catalog, profile.kind)
		if err != nil || len(fields) == 0 {
			return layoutapi.Definition{}, layoutapi.ErrInvalid
		}
		surfaces[index] = layoutapi.SurfaceDefinition{
			Kind:         profile.kind,
			Translations: defaultSurfaceTranslations(profile),
			Sections: []layoutapi.Section{{
				ID: defaultPresentationSectionID, Order: 0,
				Translations: defaultSectionTranslations(),
				Fields:       normalizeFields(fields),
			}},
		}
	}
	return layoutapi.Definition{
		SourceContract: layoutapi.DefinitionContractV2, Surfaces: surfaces,
	}, nil
}

package application

import (
	"time"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

type surfaceProfile struct {
	kind          layoutapi.SurfaceKind
	label         string
	forceReadOnly bool
}

var systemResourceSurfaceProfiles = []surfaceProfile{
	{kind: layoutapi.SurfaceRecordCreate, label: "Create"},
	{kind: layoutapi.SurfaceRecordUpdate, label: "Update"},
	{kind: layoutapi.SurfaceRecordDetail, label: "Details", forceReadOnly: true},
	{kind: layoutapi.SurfaceRecordList, label: "List", forceReadOnly: true},
	{kind: layoutapi.SurfaceRecordFilter, label: "Filters", forceReadOnly: true},
	{kind: layoutapi.SurfacePrint, label: "Print", forceReadOnly: true},
}

func compilePresentationSurfaces(
	seed string,
	createdAt time.Time,
	definition layoutapi.Definition,
	catalog presentation.FieldCatalog,
) ([]presentation.Surface, error) {
	fields, err := presentationFieldIndex(catalog)
	if err != nil {
		return nil, err
	}
	surfaces := make([]presentation.Surface, len(definition.Surfaces))
	for index, source := range definition.Surfaces {
		profile, found := profileForSurface(source.Kind)
		if !found {
			return nil, layoutapi.ErrInvalid
		}
		surface, compileErr := compilePresentationSurface(
			seed, createdAt, source, fields, profile,
		)
		if compileErr != nil {
			return nil, compileErr
		}
		surfaces[index] = surface
	}
	return surfaces, nil
}

func compilePresentationSurface(
	seed string,
	createdAt time.Time,
	source layoutapi.SurfaceDefinition,
	fields map[string]presentation.Field,
	profile surfaceProfile,
) (presentation.Surface, error) {
	surfaceID, err := deterministicUUIDv7(seed, string(profile.kind)+":surface", createdAt)
	if err != nil {
		return presentation.Surface{}, err
	}
	rootID, err := deterministicUUIDv7(seed, string(profile.kind)+":root", createdAt)
	if err != nil {
		return presentation.Surface{}, err
	}
	children, err := compilePresentationSections(
		seed, createdAt, source.Sections, fields, profile,
	)
	if err != nil {
		return presentation.Surface{}, err
	}
	return presentation.Surface{
		ID: surfaceID, Kind: string(profile.kind),
		Translations: presentationTranslations(source.Translations, profile.label),
		Root: presentation.Node{
			ID: rootID, Kind: "container", Variant: "stack", Span: 1,
			Translations: []presentation.Translation{}, Children: children,
		},
	}, nil
}

func profileForSurface(kind layoutapi.SurfaceKind) (surfaceProfile, bool) {
	for _, profile := range systemResourceSurfaceProfiles {
		if profile.kind == kind {
			return profile, true
		}
	}
	return surfaceProfile{}, false
}

func presentationFieldIndex(
	catalog presentation.FieldCatalog,
) (map[string]presentation.Field, error) {
	values, err := catalog.Fields()
	if err != nil {
		return nil, layoutapi.ErrInvalid
	}
	result := make(map[string]presentation.Field, len(values))
	for _, field := range values {
		id, idErr := field.ID()
		if idErr != nil {
			return nil, layoutapi.ErrInvalid
		}
		result[id] = field
	}
	return result, nil
}

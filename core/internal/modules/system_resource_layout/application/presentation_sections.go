package application

import (
	"time"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func compilePresentationSections(
	seed string,
	createdAt time.Time,
	sectionsInput []layoutapi.Section,
	catalog map[string]presentation.Field,
	profile surfaceProfile,
) ([]presentation.Node, error) {
	sections := make([]presentation.Node, len(sectionsInput))
	for index, section := range sectionsInput {
		id, err := deterministicUUIDv7(
			seed, string(profile.kind)+":section:"+section.ID, createdAt,
		)
		if err != nil {
			return nil, err
		}
		fields, err := compilePresentationFields(
			seed, createdAt, section.Fields, catalog, profile,
		)
		if err != nil {
			return nil, err
		}
		sections[index] = presentation.Node{
			ID: id, Kind: "container", Variant: "section", Span: 1,
			Translations: presentationTranslations(section.Translations, section.ID),
			Children:     fields,
		}
	}
	return sections, nil
}

func compilePresentationFields(
	seed string,
	createdAt time.Time,
	fields []layoutapi.Field,
	catalog map[string]presentation.Field,
	profile surfaceProfile,
) ([]presentation.Node, error) {
	result := make([]presentation.Node, len(fields))
	for index, source := range fields {
		field, exists := catalog[source.FieldID]
		if !exists {
			return nil, layoutapi.ErrInvalid
		}
		mode, err := field.Mode()
		if err != nil {
			return nil, layoutapi.ErrInvalid
		}
		interaction := "auto"
		widget := string(source.Widget)
		interaction = string(source.Interaction)
		if profile.forceReadOnly || mode == presentation.FieldReadOnly {
			if interaction != string(layoutapi.InteractionDisabled) {
				interaction = string(layoutapi.InteractionReadOnly)
			}
		}
		if profile.forceReadOnly {
			widget = "auto"
		}
		id, err := deterministicUUIDv7(
			seed, string(profile.kind)+":field:"+source.FieldID, createdAt,
		)
		if err != nil {
			return nil, err
		}
		result[index] = presentation.Node{
			ID: id, Kind: "field", FieldID: source.FieldID,
			Widget: widget, Interaction: interaction, Span: 1,
			Translations: presentationTranslations(source.Translations, source.FieldID),
		}
	}
	return result, nil
}

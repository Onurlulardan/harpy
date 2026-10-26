package application

import layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"

type legacyDefinitionWire struct {
	Contract layoutapi.DefinitionContract `json:"contract"`
	Sections []legacySectionWire          `json:"sections"`
}

type legacySectionWire struct {
	Fields       []legacyFieldWire       `json:"fields"`
	ID           string                  `json:"id"`
	Order        int32                   `json:"order"`
	Translations []layoutapi.Translation `json:"translations"`
}

type legacyFieldWire struct {
	FieldID      string                  `json:"field_id"`
	Order        int32                   `json:"order"`
	Translations []layoutapi.Translation `json:"translations"`
	Widget       layoutapi.Widget        `json:"widget"`
}

func decodeLegacyDefinition(document []byte) (layoutapi.Definition, error) {
	var wire legacyDefinitionWire
	if decodeExact(document, &wire) != nil ||
		wire.Contract != layoutapi.DefinitionContractV1 {
		return layoutapi.Definition{}, layoutapi.ErrInvalid
	}
	sections := make([]layoutapi.Section, len(wire.Sections))
	for index, source := range wire.Sections {
		fields := make([]layoutapi.Field, len(source.Fields))
		for fieldIndex, field := range source.Fields {
			fields[fieldIndex] = layoutapi.Field{
				FieldID: field.FieldID, Interaction: layoutapi.InteractionAuto,
				Order: field.Order, Translations: field.Translations, Widget: field.Widget,
			}
		}
		sections[index] = layoutapi.Section{
			ID: source.ID, Order: source.Order,
			Translations: source.Translations, Fields: fields,
		}
	}
	return legacySurfaceDefinition(sections), nil
}

func legacySurfaceDefinition(sections []layoutapi.Section) layoutapi.Definition {
	profiles := systemResourceSurfaceProfiles[:4]
	surfaces := make([]layoutapi.SurfaceDefinition, len(profiles))
	for index, profile := range profiles {
		surfaces[index] = layoutapi.SurfaceDefinition{
			Kind: profile.kind, Sections: normalizeSections(sections),
			Translations: []layoutapi.Translation{{Locale: "en", Label: profile.label}},
		}
	}
	return normalizeDefinition(layoutapi.Definition{
		SourceContract: layoutapi.DefinitionContractV1, Surfaces: surfaces,
	})
}

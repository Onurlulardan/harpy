package application

import (
	"sort"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func normalizeDefinition(input layoutapi.Definition) layoutapi.Definition {
	contract := input.SourceContract
	if contract == "" {
		contract = layoutapi.DefinitionContractV2
	}
	surfaces := make([]layoutapi.SurfaceDefinition, len(input.Surfaces))
	for index, surface := range input.Surfaces {
		surfaces[index] = layoutapi.SurfaceDefinition{
			Kind: surface.Kind, Sections: normalizeSections(surface.Sections),
			Translations: normalizeTranslations(surface.Translations),
		}
	}
	sort.Slice(surfaces, func(left, right int) bool {
		return surfaceOrder(surfaces[left].Kind) < surfaceOrder(surfaces[right].Kind)
	})
	return layoutapi.Definition{SourceContract: contract, Surfaces: surfaces}
}

func normalizeSections(input []layoutapi.Section) []layoutapi.Section {
	sections := make([]layoutapi.Section, len(input))
	for index, section := range input {
		sections[index] = layoutapi.Section{
			ID: section.ID, Order: section.Order,
			Translations: normalizeTranslations(section.Translations),
			Fields:       normalizeFields(section.Fields),
		}
	}
	sort.Slice(sections, func(left, right int) bool {
		if sections[left].Order != sections[right].Order {
			return sections[left].Order < sections[right].Order
		}
		return sections[left].ID < sections[right].ID
	})
	return sections
}

func normalizeFields(input []layoutapi.Field) []layoutapi.Field {
	fields := make([]layoutapi.Field, len(input))
	for index, field := range input {
		fields[index] = layoutapi.Field{
			FieldID: field.FieldID, Interaction: field.Interaction,
			Order: field.Order, Widget: field.Widget,
			Translations: normalizeTranslations(field.Translations),
		}
	}
	sort.Slice(fields, func(left, right int) bool {
		if fields[left].Order != fields[right].Order {
			return fields[left].Order < fields[right].Order
		}
		return fields[left].FieldID < fields[right].FieldID
	})
	return fields
}

func normalizeTranslations(input []layoutapi.Translation) []layoutapi.Translation {
	translations := append([]layoutapi.Translation(nil), input...)
	if translations == nil {
		translations = make([]layoutapi.Translation, 0)
	}
	sort.Slice(translations, func(left, right int) bool {
		return translations[left].Locale < translations[right].Locale
	})
	return translations
}

func surfaceOrder(kind layoutapi.SurfaceKind) int {
	for index, candidate := range orderedSurfaceKinds {
		if candidate == kind {
			return index
		}
	}
	return len(orderedSurfaceKinds)
}

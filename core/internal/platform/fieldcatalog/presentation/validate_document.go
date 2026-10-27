package presentation

import (
	"fmt"
	"slices"
)

type catalogField struct {
	typeID string
	mode   FieldMode
}

type validationState struct {
	document  Document
	registry  registryCatalog
	fields    map[string]catalogField
	locales   map[string]struct{}
	nodeIDs   map[string]struct{}
	nodeCount int
}

func Validate(document Document, catalog FieldCatalog) error {
	registry, err := loadRegistry(DefaultRegistry())
	if err != nil || catalog.Validate() != nil {
		return ErrInvalid
	}
	if document.RegistryContract != RegistryContractVersion ||
		len(document.Surfaces) < 1 || len(document.Surfaces) > maximumSurfaces ||
		!strictStrings(document.Locales) ||
		!slices.Contains(document.Locales, document.DefaultLocale) {
		return invalid("$.presentation", "root contract or budget")
	}
	locales := make(map[string]struct{}, len(document.Locales))
	for index, locale := range document.Locales {
		if !validLocale(locale) {
			return invalid(fmt.Sprintf("$.locales[%d]", index), "locale")
		}
		locales[locale] = struct{}{}
	}
	fieldTypes := make(map[string]catalogField, len(catalog.fields))
	for _, field := range catalog.fields {
		fieldTypes[field.id] = catalogField{typeID: string(field.typeID), mode: field.mode}
	}
	state := validationState{
		document: document, registry: registry, fields: fieldTypes,
		locales: locales, nodeIDs: make(map[string]struct{}),
	}
	return state.validateSurfaces()
}

func (state *validationState) validateSurfaces() error {
	kinds := make(map[string]struct{}, len(state.document.Surfaces))
	for index, surface := range state.document.Surfaces {
		path := fmt.Sprintf("$.surfaces[%d]", index)
		if !stableUUIDv7(surface.ID) || !validSurfaceKind(surface.Kind) {
			return invalid(path, "identity or kind")
		}
		if _, duplicate := kinds[surface.Kind]; duplicate {
			return invalid(path+".kind", "duplicate")
		}
		kinds[surface.Kind] = struct{}{}
		if err := state.validateTranslations(
			path+".translations", surface.Translations, true,
		); err != nil {
			return err
		}
		seenFields := make(map[string]struct{})
		if surface.Root.Kind != "container" || surface.Root.Span != 1 {
			return invalid(path+".root", "root container")
		}
		if err := state.validateNode(surface.Root, path+".root", "", 1, seenFields); err != nil {
			return err
		}
	}
	return nil
}

func validSurfaceKind(value string) bool {
	switch value {
	case "record_create", "record_update", "record_detail", "record_list",
		"record_filter", "portal_create", "portal_update", "portal_detail", "print":
		return true
	default:
		return false
	}
}

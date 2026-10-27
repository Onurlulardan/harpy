package application

import (
	"regexp"
	"unicode"
	"unicode/utf8"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

const (
	maximumSurfaces              = 6
	maximumSections              = 64
	maximumFields                = 512
	maximumTranslations          = 32
	maximumTranslationLabelRunes = 200
	maximumTranslationHelpRunes  = 1000
)

var orderedSurfaceKinds = []layoutapi.SurfaceKind{
	layoutapi.SurfaceRecordCreate, layoutapi.SurfaceRecordUpdate,
	layoutapi.SurfaceRecordDetail, layoutapi.SurfaceRecordList,
	layoutapi.SurfaceRecordFilter, layoutapi.SurfacePrint,
}

var localePattern = regexp.MustCompile(
	`^[a-z]{2,3}(-[A-Z][a-z]{3})?(-([A-Z]{2}|[0-9]{3}))?$`,
)

func validateDefinition(definition layoutapi.Definition) error {
	if definition.SourceContract != layoutapi.DefinitionContractV1 &&
		definition.SourceContract != layoutapi.DefinitionContractV2 ||
		len(definition.Surfaces) > maximumSurfaces {
		return layoutapi.ErrInvalid
	}
	wanted := requiredSurfaceCount(definition.SourceContract)
	seen := make(map[layoutapi.SurfaceKind]struct{}, len(definition.Surfaces))
	for index, surface := range definition.Surfaces {
		if !surface.Kind.Valid() || surfaceOrder(surface.Kind) != index ||
			validateTranslations(surface.Translations) != nil ||
			validateSurfaceSections(surface.Sections) != nil {
			return layoutapi.ErrInvalid
		}
		legacyUnsupported := surface.Kind == layoutapi.SurfaceRecordFilter ||
			surface.Kind == layoutapi.SurfacePrint
		if definition.SourceContract == layoutapi.DefinitionContractV1 && legacyUnsupported {
			return layoutapi.ErrInvalid
		}
		if _, duplicate := seen[surface.Kind]; duplicate {
			return layoutapi.ErrInvalid
		}
		seen[surface.Kind] = struct{}{}
	}
	for _, kind := range orderedSurfaceKinds[:wanted] {
		if _, exists := seen[kind]; !exists {
			return layoutapi.ErrInvalid
		}
	}
	return nil
}

func requiredSurfaceCount(contract layoutapi.DefinitionContract) int {
	if contract == layoutapi.DefinitionContractV1 {
		return 4
	}
	return 5
}

func validateSurfaceSections(sections []layoutapi.Section) error {
	if len(sections) > maximumSections {
		return layoutapi.ErrInvalid
	}
	sectionIDs := make(map[string]struct{}, len(sections))
	fieldIDs := make(map[string]struct{})
	for index, section := range sections {
		if !stableUUIDText(section.ID) || section.Order != int32(index) ||
			validateTranslations(section.Translations) != nil {
			return layoutapi.ErrInvalid
		}
		if _, duplicate := sectionIDs[section.ID]; duplicate {
			return layoutapi.ErrInvalid
		}
		sectionIDs[section.ID] = struct{}{}
		for fieldIndex, field := range section.Fields {
			if !stableUUIDText(field.FieldID) || field.Order != int32(fieldIndex) ||
				!field.Widget.Valid() || !field.Interaction.Valid() ||
				validateTranslations(field.Translations) != nil {
				return layoutapi.ErrInvalid
			}
			if _, duplicate := fieldIDs[field.FieldID]; duplicate {
				return layoutapi.ErrInvalid
			}
			fieldIDs[field.FieldID] = struct{}{}
			if len(fieldIDs) > maximumFields {
				return layoutapi.ErrInvalid
			}
		}
	}
	return nil
}

func validateTranslations(translations []layoutapi.Translation) error {
	if len(translations) > maximumTranslations {
		return layoutapi.ErrInvalid
	}
	previous := ""
	for _, translation := range translations {
		if !localePattern.MatchString(translation.Locale) ||
			(previous != "" && previous >= translation.Locale) ||
			!validPresentationText(translation.Label, 1, maximumTranslationLabelRunes) ||
			!validPresentationText(translation.Help, 0, maximumTranslationHelpRunes) {
			return layoutapi.ErrInvalid
		}
		previous = translation.Locale
	}
	return nil
}

func validPresentationText(value string, minimum, maximum int) bool {
	if !utf8.ValidString(value) {
		return false
	}
	count := utf8.RuneCountInString(value)
	if count < minimum || count > maximum {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func stableUUID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}

func stableUUIDText(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && stableUUID(parsed) && parsed.String() == value
}

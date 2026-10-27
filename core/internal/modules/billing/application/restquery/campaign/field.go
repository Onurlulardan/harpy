package campaign

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func buildField(spec fieldSpec) (api.FieldDefinition, error) {
	kind, kindErr := fieldType(spec)
	storage, storageErr := fieldStorage(spec)
	filters := make([]api.FilterCapabilityInput, len(spec.Filters))
	for index, name := range spec.Filters {
		operator, found := api.ParseOperator(name)
		if !found {
			return api.FieldDefinition{}, ErrContract
		}
		filters[index] = api.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	capabilities, capabilityErr := api.NewFieldCapabilities(kind, api.FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Sortable: spec.Sortable,
		CursorCanonicalBytes: spec.Cursor, Searchable: spec.Search,
	})
	if kindErr != nil || storageErr != nil || capabilityErr != nil {
		return api.FieldDefinition{}, ErrContract
	}
	field, err := api.NewFieldDefinition(spec.ID, spec.Name, kind, storage, capabilities)
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	return field, nil
}

func fieldType(spec fieldSpec) (api.FieldType, error) {
	switch spec.Kind {
	case "uuid":
		return api.NewUUIDFieldType()
	case "timestamp":
		return api.NewTimestampFieldType()
	case "text", "search_text":
		casefold := api.TextCasefoldDisabled
		if spec.Kind == "search_text" {
			casefold = api.TextCasefoldUnicodeDefaultV1
		}
		profile, err := api.NewTextProfile(casefold, api.TextAccentDisabled)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		return api.NewTextFieldType(profile)
	case "integer":
		minimum, maximum := int64(1), int64(math.MaxInt64)
		if spec.Profile == "signed" {
			minimum, maximum = math.MinInt32, math.MaxInt32
		} else if spec.Profile == "count" {
			minimum, maximum = 0, 256
		}
		profile, err := api.NewIntegerProfile(minimum, maximum)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		return api.NewIntegerFieldType(profile)
	case "enum":
		return enumType(spec.Profile)
	default:
		return api.FieldType{}, ErrContract
	}
}

func enumType(profile string) (api.FieldType, error) {
	version, codes := lifecycleEnumID, []string{"active", "retired"}
	if profile == "publication" {
		version, codes = publicationEnumID, []string{"published"}
	} else if profile != "lifecycle" {
		return api.FieldType{}, ErrContract
	}
	values := make([]api.EnumValue, len(codes))
	for index, code := range codes {
		value, err := api.NewEnumValue(code, nil)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		values[index] = value
	}
	enum, err := api.NewEnumProfile(version, values)
	if err != nil {
		return api.FieldType{}, ErrContract
	}
	return api.NewEnumFieldType(enum)
}

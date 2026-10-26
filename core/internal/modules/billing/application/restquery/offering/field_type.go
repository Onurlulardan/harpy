package offering

import (
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func fieldType(spec fieldSpec) (api.FieldType, error) {
	switch spec.Kind {
	case "uuid":
		return api.NewUUIDFieldType()
	case "boolean":
		return api.NewBooleanFieldType()
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
		minimum, maximum := integerBounds(spec.Profile)
		profile, err := api.NewIntegerProfile(minimum, maximum)
		if err != nil {
			return api.FieldType{}, ErrContract
		}
		return api.NewIntegerFieldType(profile)
	case "enum":
		return enumFieldType(spec.Profile)
	default:
		return api.FieldType{}, ErrContract
	}
}

func integerBounds(profile string) (int64, int64) {
	switch profile {
	case "days":
		return 0, 3650
	case "price_count":
		return 1, 16
	case "grant_count":
		return 0, 256
	default:
		return 1, math.MaxInt64
	}
}

func enumFieldType(profile string) (api.FieldType, error) {
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

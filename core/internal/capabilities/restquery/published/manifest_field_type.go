package published

import (
	"fmt"
	"math"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func decodeParameterFieldType(wire parameterWire) (api.FieldType, error) {
	switch wire.Type {
	case "text":
		casefold, accent, err := decodeTextProfiles(wire.Casefold, wire.Accent)
		if err != nil {
			return api.FieldType{}, err
		}
		profile, err := api.NewTextProfile(casefold, accent)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewTextFieldType(profile)
	case "integer":
		minimum, maximum := int64(math.MinInt64), int64(math.MaxInt64)
		if wire.Minimum != nil {
			minimum = *wire.Minimum
		}
		if wire.Maximum != nil {
			maximum = *wire.Maximum
		}
		profile, err := api.NewIntegerProfile(minimum, maximum)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewIntegerFieldType(profile)
	case "decimal":
		if wire.Precision == nil || wire.Scale == nil {
			return api.FieldType{}, fmt.Errorf("decimal profile is required")
		}
		profile, err := api.NewDecimalProfile(*wire.Precision, *wire.Scale)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewDecimalFieldType(profile)
	case "boolean":
		return api.NewBooleanFieldType()
	case "date":
		return api.NewDateFieldType()
	case "timestamp":
		return api.NewTimestampFieldType()
	case "uuid":
		return api.NewUUIDFieldType()
	case "relation":
		return api.NewRelationFieldType()
	case "json":
		return api.NewJSONFieldType()
	case "enum":
		values := make([]api.EnumValue, len(wire.EnumValues))
		for index, value := range wire.EnumValues {
			decoded, err := api.NewEnumValue(value.Code, value.Ordinal)
			if err != nil {
				return api.FieldType{}, err
			}
			values[index] = decoded
		}
		profile, err := api.NewEnumProfile(wire.EnumVersion, values)
		if err != nil {
			return api.FieldType{}, err
		}
		return api.NewEnumFieldType(profile)
	default:
		return api.FieldType{}, fmt.Errorf("published parameter type is invalid")
	}
}

func decodeTextProfiles(
	casefoldWire, accentWire string,
) (api.TextCasefoldProfile, api.TextAccentProfile, error) {
	casefold := api.TextCasefoldDisabled
	if casefoldWire == "unicode_default_casefold_v1" {
		casefold = api.TextCasefoldUnicodeDefaultV1
	} else if casefoldWire != "" && casefoldWire != "none" {
		return 0, 0, fmt.Errorf("text casefold profile is invalid")
	}
	accent := api.TextAccentDisabled
	if accentWire == "accent_insensitive_v1" {
		accent = api.TextAccentInsensitiveV1
	} else if accentWire != "" && accentWire != "none" {
		return 0, 0, fmt.Errorf("text accent profile is invalid")
	}
	return casefold, accent, nil
}

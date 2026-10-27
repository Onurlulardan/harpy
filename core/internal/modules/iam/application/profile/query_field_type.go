package profile

import restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func queryFieldType(field Field) (restquery.FieldType, error) {
	if field.cardinality == CardinalityList {
		if field.kind == LogicalString {
			return restquery.NewTextListFieldType(exactQueryTextProfile())
		}
		if field.kind == LogicalUUID {
			return restquery.NewUUIDListFieldType()
		}
		return restquery.FieldType{}, ErrQueryCapabilityRejected
	}
	switch field.kind {
	case LogicalString:
		return restquery.NewTextFieldType(exactQueryTextProfile())
	case LogicalInteger:
		profile, err := restquery.NewIntegerProfile(
			field.validation.integerMinimum, field.validation.integerMaximum,
		)
		if err != nil {
			return restquery.FieldType{}, err
		}
		return restquery.NewIntegerFieldType(profile)
	case LogicalDecimal:
		profile, err := restquery.NewDecimalProfile(
			field.validation.decimalPrecision, field.validation.decimalScale,
		)
		if err != nil {
			return restquery.FieldType{}, err
		}
		return restquery.NewDecimalFieldType(profile)
	case LogicalBoolean:
		return restquery.NewBooleanFieldType()
	case LogicalDate:
		return restquery.NewDateFieldType()
	case LogicalTimestamp:
		return restquery.NewTimestampFieldType()
	case LogicalUUID:
		return restquery.NewUUIDFieldType()
	case LogicalEnum:
		return queryEnumFieldType(field)
	case LogicalJSON:
		return restquery.NewJSONFieldType()
	default:
		return restquery.FieldType{}, ErrQueryCapabilityRejected
	}
}

func exactQueryTextProfile() restquery.TextProfile {
	profile, _ := restquery.NewTextProfile(
		restquery.TextCasefoldDisabled, restquery.TextAccentDisabled,
	)
	return profile
}

func queryEnumFieldType(field Field) (restquery.FieldType, error) {
	values := make([]restquery.EnumValue, len(field.validation.enumValues))
	for index, code := range field.validation.enumValues {
		value, err := restquery.NewEnumValue(code, nil)
		if err != nil {
			return restquery.FieldType{}, err
		}
		values[index] = value
	}
	profile, err := restquery.NewEnumProfile(field.id.String(), values)
	if err != nil {
		return restquery.FieldType{}, err
	}
	return restquery.NewEnumFieldType(profile)
}

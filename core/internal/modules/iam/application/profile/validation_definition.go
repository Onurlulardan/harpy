package profile

import "strconv"

type validationDefinition struct {
	MaximumBytes     int      `json:"maximum_bytes"`
	MaximumItems     int      `json:"maximum_items,omitempty"`
	IntegerMinimum   string   `json:"integer_minimum,omitempty"`
	IntegerMaximum   string   `json:"integer_maximum,omitempty"`
	DecimalPrecision int      `json:"decimal_precision,omitempty"`
	DecimalScale     *int     `json:"decimal_scale,omitempty"`
	EnumValues       []string `json:"enum_values,omitempty"`
}

func (value validation) definition(kind LogicalType) validationDefinition {
	definition := validationDefinition{
		MaximumBytes: value.maximumBytes, MaximumItems: value.maximumItems,
	}
	switch kind {
	case LogicalInteger:
		definition.IntegerMinimum = strconv.FormatInt(value.integerMinimum, 10)
		definition.IntegerMaximum = strconv.FormatInt(value.integerMaximum, 10)
	case LogicalDecimal:
		definition.DecimalPrecision = value.decimalPrecision
		scale := value.decimalScale
		definition.DecimalScale = &scale
	case LogicalEnum:
		definition.EnumValues = append([]string(nil), value.enumValues...)
	}
	return definition
}

func (value validation) permitsEnum(candidate string) bool {
	for _, allowed := range value.enumValues {
		if allowed == candidate {
			return true
		}
	}
	return false
}

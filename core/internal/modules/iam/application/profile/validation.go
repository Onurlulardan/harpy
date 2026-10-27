package profile

import (
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type ValidationInput struct {
	MaximumBytes     int
	MaximumItems     int
	IntegerMinimum   *int64
	IntegerMaximum   *int64
	DecimalPrecision int
	DecimalScale     int
	EnumValues       []string
}

type validation struct {
	maximumBytes     int
	maximumItems     int
	integerMinimum   int64
	integerMaximum   int64
	decimalPrecision int
	decimalScale     int
	enumValues       []string
}

func newValidation(
	kind LogicalType, cardinality Cardinality, input ValidationInput,
) (validation, error) {
	result := validation{maximumBytes: input.MaximumBytes}
	if result.maximumBytes == 0 {
		result.maximumBytes = MaximumDocumentBytes
	}
	if result.maximumBytes < 1 || result.maximumBytes > MaximumDocumentBytes {
		return validation{}, ErrValidationInvalid
	}
	if err := result.configureCardinality(cardinality, input.MaximumItems); err != nil {
		return validation{}, err
	}
	if err := result.configureType(kind, input); err != nil {
		return validation{}, err
	}
	return result, nil
}

func (value *validation) configureCardinality(
	cardinality Cardinality, maximumItems int,
) error {
	if cardinality == CardinalitySingle {
		if maximumItems != 0 {
			return ErrValidationInvalid
		}
		return nil
	}
	if maximumItems == 0 {
		maximumItems = MaximumListItems
	}
	if maximumItems < 1 || maximumItems > MaximumListItems {
		return ErrValidationInvalid
	}
	value.maximumItems = maximumItems
	return nil
}

func (value *validation) configureType(kind LogicalType, input ValidationInput) error {
	if kind == LogicalInteger {
		return value.configureInteger(input)
	}
	if input.IntegerMinimum != nil || input.IntegerMaximum != nil {
		return ErrValidationInvalid
	}
	if kind == LogicalDecimal {
		return value.configureDecimal(input)
	}
	if input.DecimalPrecision != 0 || input.DecimalScale != 0 {
		return ErrValidationInvalid
	}
	if kind == LogicalEnum {
		return value.configureEnum(input.EnumValues)
	}
	if input.EnumValues != nil {
		return ErrValidationInvalid
	}
	return nil
}

func (value *validation) configureInteger(input ValidationInput) error {
	value.integerMinimum, value.integerMaximum = math.MinInt64, math.MaxInt64
	if input.IntegerMinimum != nil {
		value.integerMinimum = *input.IntegerMinimum
	}
	if input.IntegerMaximum != nil {
		value.integerMaximum = *input.IntegerMaximum
	}
	if value.integerMinimum > value.integerMaximum || input.DecimalPrecision != 0 ||
		input.DecimalScale != 0 || input.EnumValues != nil {
		return ErrValidationInvalid
	}
	return nil
}

func (value *validation) configureDecimal(input ValidationInput) error {
	if input.DecimalPrecision < 1 || input.DecimalPrecision > 1000 ||
		input.DecimalScale < 0 || input.DecimalScale > input.DecimalPrecision ||
		input.EnumValues != nil {
		return ErrValidationInvalid
	}
	value.decimalPrecision = input.DecimalPrecision
	value.decimalScale = input.DecimalScale
	return nil
}

func (value *validation) configureEnum(values []string) error {
	if len(values) < 1 || len(values) > MaximumEnumValues {
		return ErrValidationInvalid
	}
	seen := make(map[string]struct{}, len(values))
	value.enumValues = append([]string(nil), values...)
	for _, candidate := range value.enumValues {
		if candidate == "" || !utf8.ValidString(candidate) ||
			strings.ContainsRune(candidate, '\x00') || !norm.NFC.IsNormalString(candidate) {
			return ErrValidationInvalid
		}
		if _, duplicate := seen[candidate]; duplicate {
			return ErrValidationInvalid
		}
		seen[candidate] = struct{}{}
	}
	sort.Strings(value.enumValues)
	return nil
}

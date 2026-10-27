package profile

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
)

type fieldValidationWire struct {
	MaximumBytes     *int      `json:"maximum_bytes"`
	MaximumItems     *int      `json:"maximum_items"`
	IntegerMinimum   *string   `json:"integer_minimum"`
	IntegerMaximum   *string   `json:"integer_maximum"`
	DecimalPrecision *int      `json:"decimal_precision"`
	DecimalScale     *int      `json:"decimal_scale"`
	EnumValues       *[]string `json:"enum_values"`
}

func decodeFieldValidation(document []byte) (ValidationInput, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var wire fieldValidationWire
	if err := decoder.Decode(&wire); err != nil || wire.MaximumBytes == nil {
		return ValidationInput{}, ErrValidationInvalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ValidationInput{}, ErrValidationInvalid
	}
	result := ValidationInput{MaximumBytes: *wire.MaximumBytes}
	if wire.MaximumItems != nil {
		result.MaximumItems = *wire.MaximumItems
	}
	if wire.DecimalPrecision != nil {
		result.DecimalPrecision = *wire.DecimalPrecision
	}
	if wire.DecimalScale != nil {
		result.DecimalScale = *wire.DecimalScale
	}
	if wire.EnumValues != nil {
		result.EnumValues = append([]string(nil), (*wire.EnumValues)...)
	}
	minimum, err := decodeSnapshotInteger(wire.IntegerMinimum)
	if err != nil {
		return ValidationInput{}, err
	}
	maximum, err := decodeSnapshotInteger(wire.IntegerMaximum)
	if err != nil {
		return ValidationInput{}, err
	}
	result.IntegerMinimum, result.IntegerMaximum = minimum, maximum
	return result, nil
}

func decodeSnapshotInteger(value *string) (*int64, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(*value, 10, 64)
	if err != nil || strconv.FormatInt(parsed, 10) != *value {
		return nil, ErrValidationInvalid
	}
	return &parsed, nil
}

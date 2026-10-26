package maintenance

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

type profileValidationWire struct {
	MaximumBytes     int      `json:"maximum_bytes"`
	MaximumItems     int      `json:"maximum_items"`
	IntegerMinimum   string   `json:"integer_minimum"`
	IntegerMaximum   string   `json:"integer_maximum"`
	DecimalPrecision int      `json:"decimal_precision"`
	DecimalScale     *int     `json:"decimal_scale"`
	EnumValues       []string `json:"enum_values"`
}

func decodeProfileValidation(
	document []byte,
	kind profileapp.LogicalType,
) (profileapp.ValidationInput, error) {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var wire profileValidationWire
	if decoder.Decode(&wire) != nil {
		return profileapp.ValidationInput{}, ErrRuntimeInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return profileapp.ValidationInput{}, ErrRuntimeInvalid
	}
	result := profileapp.ValidationInput{
		MaximumBytes: wire.MaximumBytes, MaximumItems: wire.MaximumItems,
		DecimalPrecision: wire.DecimalPrecision, EnumValues: wire.EnumValues,
	}
	if wire.DecimalScale != nil {
		result.DecimalScale = *wire.DecimalScale
	}
	if kind == profileapp.LogicalInteger {
		minimum, minimumErr := strconv.ParseInt(wire.IntegerMinimum, 10, 64)
		maximum, maximumErr := strconv.ParseInt(wire.IntegerMaximum, 10, 64)
		if minimumErr != nil || maximumErr != nil {
			return profileapp.ValidationInput{}, ErrRuntimeInvalid
		}
		result.IntegerMinimum, result.IntegerMaximum = &minimum, &maximum
	}
	return result, nil
}

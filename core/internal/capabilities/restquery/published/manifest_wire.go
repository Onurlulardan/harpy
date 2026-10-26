package published

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type manifestWire struct {
	Contract        string          `json:"contract"`
	MaximumPageSize int             `json:"maximum_page_size"`
	Parameters      []parameterWire `json:"parameters"`
}

type parameterWire struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Arity       string          `json:"arity"`
	Required    bool            `json:"required"`
	Default     json.RawMessage `json:"default,omitempty"`
	Minimum     *int64          `json:"minimum,omitempty"`
	Maximum     *int64          `json:"maximum,omitempty"`
	Precision   *int            `json:"precision,omitempty"`
	Scale       *int            `json:"scale,omitempty"`
	Casefold    string          `json:"casefold,omitempty"`
	Accent      string          `json:"accent,omitempty"`
	EnumVersion string          `json:"enum_version,omitempty"`
	EnumValues  []enumValueWire `json:"enum_values,omitempty"`
}

type enumValueWire struct {
	Code    string `json:"code"`
	Ordinal *int   `json:"ordinal,omitempty"`
}

func DecodeParameterManifest(document []byte) (ParameterManifest, error) {
	if bytes.Equal(bytes.TrimSpace(document), []byte(`{}`)) {
		return NewParameterManifest(nil, 50)
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var wire manifestWire
	if err := decoder.Decode(&wire); err != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		wire.Contract != "restquery-parameters.v1" || wire.MaximumPageSize < 1 ||
		wire.Parameters == nil {
		return ParameterManifest{}, fmt.Errorf("published parameter manifest document is invalid")
	}
	inputs := make([]ParameterInput, len(wire.Parameters))
	for index, parameter := range wire.Parameters {
		fieldType, err := decodeParameterFieldType(parameter)
		arity, arityErr := decodeParameterArity(parameter.Arity)
		if err != nil || arityErr != nil {
			return ParameterManifest{}, fmt.Errorf("published parameter manifest document is invalid")
		}
		defaults, err := decodeParameterDefaults(parameter.Default, fieldType, arity)
		if err != nil {
			return ParameterManifest{}, fmt.Errorf("published parameter manifest document is invalid")
		}
		inputs[index] = ParameterInput{Name: parameter.Name, Type: fieldType,
			Arity: arity, Required: parameter.Required, Default: defaults}
	}
	return NewParameterManifest(inputs, wire.MaximumPageSize)
}

func decodeParameterArity(value string) (ParameterArity, error) {
	switch value {
	case "single":
		return ParameterAritySingle, nil
	case "list":
		return ParameterArityList, nil
	default:
		return 0, fmt.Errorf("published parameter arity is invalid")
	}
}

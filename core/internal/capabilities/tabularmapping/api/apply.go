package tabularmapping

import (
	"encoding/json"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func (program Program) Apply(row []byte) ([]byte, error) {
	if len(program.fields) < 1 || len(program.document) < 2 {
		return nil, ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(row, MaximumRowBytes)
	if err != nil {
		return nil, ErrInvalid
	}
	var source map[string]json.RawMessage
	if json.Unmarshal(canonical, &source) != nil {
		return nil, ErrInvalid
	}
	target := make(map[string]json.RawMessage, len(program.fields))
	for _, field := range program.fields {
		value, present, err := field.apply(source)
		if err != nil {
			return nil, err
		}
		if present {
			target[field.target] = value
		}
	}
	encoded, err := json.Marshal(target)
	if err != nil {
		return nil, ErrInvalid
	}
	return canonicaljsonv1.CanonicalizeObject(encoded, MaximumRowBytes)
}

func (field compiledField) apply(
	source map[string]json.RawMessage,
) (json.RawMessage, bool, error) {
	switch field.kind {
	case "direct":
		value, found := source[field.source]
		return field.require(value, found && string(value) != "null")
	case "concat":
		parts := make([]string, len(field.sources))
		for index, name := range field.sources {
			value, found := source[name]
			if !found || string(value) == "null" {
				return field.require(nil, false)
			}
			part, err := scalarText(value)
			if err != nil {
				return nil, false, err
			}
			parts[index] = part
		}
		encoded, _ := json.Marshal(strings.Join(parts, field.separator))
		return encoded, true, nil
	case "static":
		return append(json.RawMessage(nil), field.static...), true, nil
	case "formula_numeric":
		value, missing, err := evaluateNumeric(*field.numeric, source)
		if err != nil {
			return nil, false, err
		}
		if missing {
			return field.require(nil, false)
		}
		encoded, _ := json.Marshal(value)
		return encoded, true, nil
	default:
		return nil, false, ErrInvalid
	}
}

func (field compiledField) require(
	value json.RawMessage,
	present bool,
) (json.RawMessage, bool, error) {
	if present {
		return append(json.RawMessage(nil), value...), true, nil
	}
	if field.required {
		return nil, false, ErrRequiredMissing
	}
	return nil, false, nil
}

func scalarText(raw json.RawMessage) (string, error) {
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return "", ErrInvalid
	}
	switch typed := value.(type) {
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	default:
		return "", ErrInvalid
	}
}

package tabularmapping

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

type expressionWire struct {
	Kind      string          `json:"kind"`
	Source    string          `json:"source,omitempty"`
	Sources   []string        `json:"sources,omitempty"`
	Separator string          `json:"separator,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	AST       *numericNode    `json:"ast,omitempty"`
}

type compiledField struct {
	target, kind, source, separator string
	required                        bool
	sources                         []string
	static                          json.RawMessage
	numeric                         *numericNode
}

func compileExpression(field mappingField) (compiledField, int, error) {
	decoder := json.NewDecoder(bytes.NewReader(field.Expression))
	decoder.DisallowUnknownFields()
	var wire expressionWire
	if decoder.Decode(&wire) != nil {
		return compiledField{}, 0, ErrInvalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return compiledField{}, 0, ErrInvalid
	}
	compiled := compiledField{target: field.Target, required: field.Required,
		kind: wire.Kind, source: wire.Source, separator: wire.Separator,
		sources: append([]string(nil), wire.Sources...),
		static:  append(json.RawMessage(nil), wire.Value...), numeric: wire.AST}
	switch wire.Kind {
	case "direct":
		if !validSource(wire.Source) || wire.Sources != nil || wire.Separator != "" ||
			wire.Value != nil || wire.AST != nil {
			return compiledField{}, 0, ErrInvalid
		}
	case "concat":
		if wire.Source != "" || wire.Value != nil || wire.AST != nil ||
			len(wire.Sources) < 1 || len(wire.Sources) > MaximumConcatSources ||
			len(wire.Separator) > MaximumSourceBytes || !utf8.ValidString(wire.Separator) {
			return compiledField{}, 0, ErrInvalid
		}
		for _, source := range wire.Sources {
			if !validSource(source) {
				return compiledField{}, 0, ErrInvalid
			}
		}
	case "static":
		if wire.Source != "" || wire.Sources != nil || wire.Separator != "" ||
			wire.AST != nil || len(wire.Value) < 1 || len(wire.Value) > MaximumStaticBytes {
			return compiledField{}, 0, ErrInvalid
		}
	case "formula_numeric":
		if wire.Source != "" || wire.Sources != nil || wire.Separator != "" ||
			wire.Value != nil || wire.AST == nil {
			return compiledField{}, 0, ErrInvalid
		}
		nodes, err := validateNumericNode(*wire.AST, 1)
		return compiled, nodes, err
	default:
		return compiledField{}, 0, ErrInvalid
	}
	return compiled, 0, nil
}

func validSource(value string) bool {
	return value != "" && len(value) <= MaximumSourceBytes && utf8.ValidString(value)
}

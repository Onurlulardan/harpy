package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
)

const queryValueRedaction = "[REDACTED]"
const maximumLiteralBytes = 64 * 1024

type literalRepresentation uint8

const (
	literalRepresentationJSON literalRepresentation = iota + 1
	literalRepresentationText
)

// Literal retains either a POST JSON value or a decoded GET text leaf.
// Catalog resolution is the only layer that assigns its logical type.
type Literal struct {
	representation literalRepresentation
	json           []byte
	text           string
	validated      bool
}

func NewJSONLiteral(value []byte) (Literal, error) {
	if len(value) == 0 || len(value) > maximumLiteralBytes ||
		!utf8.Valid(value) || !json.Valid(value) {
		return Literal{}, fmt.Errorf("REST Query literal must be one valid UTF-8 JSON value")
	}
	trimmed := bytes.Trim(value, " \t\r\n")
	scan, scanErr := jsonscan.Scan(trimmed, jsonscan.MaximumNesting)
	if scanErr != nil || len(scan.DuplicatePointers()) != 0 ||
		bytes.Equal(trimmed, []byte("null")) {
		return Literal{}, fmt.Errorf("REST Query literal is invalid")
	}
	literal := Literal{
		representation: literalRepresentationJSON,
		json:           bytes.Clone(trimmed), validated: true,
	}
	if err := literal.Validate(); err != nil {
		return Literal{}, err
	}
	return literal, nil
}

func NewTextLiteral(value string) (Literal, error) {
	if len(value) > maximumLiteralBytes || !utf8.ValidString(value) ||
		strings.ContainsRune(value, '\x00') {
		return Literal{}, fmt.Errorf("REST Query text literal is invalid")
	}
	literal := Literal{
		representation: literalRepresentationText, text: value, validated: true,
	}
	if err := literal.Validate(); err != nil {
		return Literal{}, err
	}
	return literal, nil
}

func (literal Literal) JSONValue() ([]byte, bool, error) {
	if err := literal.Validate(); err != nil {
		return nil, false, err
	}
	return bytes.Clone(literal.json), literal.representation == literalRepresentationJSON, nil
}

func (literal Literal) TextValue() (string, bool, error) {
	if err := literal.Validate(); err != nil {
		return "", false, err
	}
	return literal.text, literal.representation == literalRepresentationText, nil
}

func (literal Literal) String() string   { return queryValueRedaction }
func (literal Literal) GoString() string { return queryValueRedaction }
func (Literal) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
func (literal Literal) Format(state fmt.State, _ rune) { writeQueryRedaction(state) }

func (literal Literal) clone() Literal {
	result := literal
	result.json = bytes.Clone(literal.json)
	return result
}

func (literal Literal) valid() bool { return literal.Validate() == nil }

func (literal Literal) Validate() error {
	valid := literal.validated
	switch literal.representation {
	case literalRepresentationJSON:
		scan, scanErr := jsonscan.Scan(literal.json, jsonscan.MaximumNesting)
		valid = valid && scanErr == nil && len(scan.DuplicatePointers()) == 0 &&
			literal.text == "" && len(literal.json) > 0 &&
			len(literal.json) <= maximumLiteralBytes &&
			utf8.Valid(literal.json) && json.Valid(literal.json) &&
			!bytes.Equal(literal.json, []byte("null")) &&
			bytes.Equal(literal.json, bytes.Trim(literal.json, " \t\r\n"))
	case literalRepresentationText:
		valid = valid && literal.json == nil && len(literal.text) <= maximumLiteralBytes &&
			utf8.ValidString(literal.text) &&
			!strings.ContainsRune(literal.text, '\x00')
	default:
		valid = false
	}
	if !valid {
		return fmt.Errorf("REST Query literal is invalid")
	}
	return nil
}

func writeQueryRedaction(state fmt.State) {
	_, _ = state.Write([]byte(queryValueRedaction))
}

func rejectDirectJSONMarshal() ([]byte, error) {
	return nil, fmt.Errorf("REST Query API values cannot be marshaled directly")
}

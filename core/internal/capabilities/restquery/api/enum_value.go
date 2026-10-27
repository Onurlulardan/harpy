package api

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type EnumValue struct {
	code       string
	ordinal    int
	hasOrdinal bool
	validated  bool
}

func NewEnumValue(code string, ordinal *int) (EnumValue, error) {
	value := EnumValue{code: code, validated: true}
	if ordinal != nil {
		value.ordinal, value.hasOrdinal = *ordinal, true
	}
	if err := value.Validate(); err != nil {
		return EnumValue{}, err
	}
	return value, nil
}

func (value EnumValue) Code() (string, error) {
	if err := value.Validate(); err != nil {
		return "", err
	}
	return value.code, nil
}

func (value EnumValue) Ordinal() (int, bool, error) {
	if err := value.Validate(); err != nil {
		return 0, false, err
	}
	return value.ordinal, value.hasOrdinal, nil
}

func (value EnumValue) Validate() error {
	if !value.validated || value.code == "" || !utf8.ValidString(value.code) ||
		strings.ContainsRune(value.code, '\x00') || !norm.NFC.IsNormalString(value.code) {
		return fmt.Errorf("REST Query enum value is invalid")
	}
	return nil
}

func (value EnumValue) String() string   { return queryValueRedaction }
func (value EnumValue) GoString() string { return queryValueRedaction }
func (value EnumValue) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (EnumValue) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

package resolved

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// OutputKey is the only presentation name retained by a resolved query.
type OutputKey struct {
	value     string
	validated bool
}

func NewOutputKey(value string) (OutputKey, error) {
	key := OutputKey{value: value, validated: true}
	if err := key.Validate(); err != nil {
		return OutputKey{}, err
	}
	return key, nil
}

func (key OutputKey) Validate() error {
	if !key.validated || key.value == "" || key.value == "_included" ||
		!utf8.ValidString(key.value) || strings.ContainsRune(key.value, '\x00') ||
		!norm.NFC.IsNormalString(key.value) {
		return fmt.Errorf("resolved REST Query output key is invalid")
	}
	return nil
}

func (key OutputKey) Value() (string, error) {
	if err := key.Validate(); err != nil {
		return "", err
	}
	return key.value, nil
}

func (OutputKey) String() string                     { return valueRedaction }
func (OutputKey) GoString() string                   { return valueRedaction }
func (key OutputKey) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (OutputKey) MarshalJSON() ([]byte, error)       { return rejectDirectJSONMarshal() }

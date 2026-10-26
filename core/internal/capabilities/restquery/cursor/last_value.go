package cursor

import (
	"fmt"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type LastValue struct {
	value resolved.Value
	null  bool
	valid bool
}

func NewLastValue(value resolved.Value) (LastValue, error) {
	result := LastValue{value: value, valid: true}
	if err := result.Validate(); err != nil {
		return LastValue{}, err
	}
	return result, nil
}

func NewNullLastValue() LastValue {
	return LastValue{null: true, valid: true}
}

func (value LastValue) Validate() error {
	if !value.valid || value.null && value.value.Validate() == nil ||
		!value.null && value.value.Validate() != nil {
		return errInvalidLastValue
	}
	return nil
}

func (value LastValue) IsNull() (bool, error) {
	if err := value.Validate(); err != nil {
		return false, err
	}
	return value.null, nil
}

func (value LastValue) Value() (resolved.Value, bool, error) {
	if err := value.Validate(); err != nil {
		return resolved.Value{}, false, err
	}
	return value.value, !value.null, nil
}

func (LastValue) String() string                       { return redacted }
func (LastValue) GoString() string                     { return redacted }
func (LastValue) LogValue() slog.Value                 { return redactedLogValue() }
func (value LastValue) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (LastValue) MarshalJSON() ([]byte, error)         { return rejectJSONMarshal() }

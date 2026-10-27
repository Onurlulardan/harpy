package rowdecode

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

// Value is one validated PostgreSQL value in canonical REST response form.
type Value struct {
	fieldType api.FieldType
	kind      api.LogicalType
	response  []byte
	last      cursor.LastValue
	hasLast   bool
	null      bool
	validated bool
}

func newValue(
	fieldType api.FieldType,
	kind api.LogicalType,
	response []byte,
	last cursor.LastValue,
	hasLast bool,
	isNull bool,
) (Value, error) {
	clonedType, err := fieldType.Clone()
	if err != nil {
		return Value{}, ErrInvalidFieldType
	}
	value := Value{
		fieldType: clonedType, kind: kind, response: bytes.Clone(response), last: last,
		hasLast: hasLast, null: isNull, validated: true,
	}
	if err := value.Validate(); err != nil {
		return Value{}, err
	}
	return value, nil
}

func (value Value) Validate() error {
	responseNull := bytes.Equal(value.response, []byte("null"))
	declaredKind, typeErr := value.fieldType.Kind()
	if !value.validated || typeErr != nil || declaredKind != value.kind ||
		!value.kind.Valid() || len(value.response) == 0 ||
		value.hasLast && value.last.Validate() != nil ||
		!value.hasLast && value.last.Validate() == nil ||
		value.null && !responseNull ||
		!value.null && responseNull && value.kind != api.LogicalTypeJSON {
		return ErrInvalidValue
	}
	return nil
}

func (value Value) FieldType() (api.FieldType, error) {
	if err := value.Validate(); err != nil {
		return api.FieldType{}, err
	}
	fieldType, err := value.fieldType.Clone()
	if err != nil {
		return api.FieldType{}, ErrInvalidValue
	}
	return fieldType, nil
}

func (value Value) Kind() (api.LogicalType, error) {
	if err := value.Validate(); err != nil {
		return 0, err
	}
	return value.kind, nil
}

func (value Value) IsNull() (bool, error) {
	if err := value.Validate(); err != nil {
		return false, err
	}
	return value.null, nil
}

func (value Value) ResponseJSON() ([]byte, error) {
	if err := value.Validate(); err != nil {
		return nil, err
	}
	return bytes.Clone(value.response), nil
}

func (value Value) LastValue() (cursor.LastValue, bool, error) {
	if err := value.Validate(); err != nil {
		return cursor.LastValue{}, false, err
	}
	return value.last, value.hasLast, nil
}

func (Value) String() string                       { return redacted }
func (Value) GoString() string                     { return redacted }
func (Value) LogValue() slog.Value                 { return slog.StringValue(redacted) }
func (value Value) Format(state fmt.State, _ rune) { _, _ = io.WriteString(state, redacted) }
func (Value) MarshalJSON() ([]byte, error)         { return nil, ErrDirectMarshal }

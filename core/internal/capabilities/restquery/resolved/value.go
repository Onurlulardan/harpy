package resolved

import (
	"bytes"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/internal/jsonscan"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type valueState uint8

const (
	valueStateScalar valueState = iota + 1
	valueStateBoolean
	valueStateJSON
)

// Value is a catalog-resolved literal. Its formatting surfaces never reveal data.
type Value struct {
	kind      api.LogicalType
	state     valueState
	scalar    string
	boolean   bool
	json      []byte
	validated bool
}

func NewScalarValue(kind api.LogicalType, canonical string) (Value, error) {
	value := Value{
		kind: kind, state: valueStateScalar, scalar: canonical, validated: true,
	}
	if err := value.Validate(); err != nil {
		return Value{}, err
	}
	return value, nil
}

func NewBooleanValue(value bool) (Value, error) {
	result := Value{
		kind: api.LogicalTypeBoolean, state: valueStateBoolean,
		boolean: value, validated: true,
	}
	if err := result.Validate(); err != nil {
		return Value{}, err
	}
	return result, nil
}

func NewJSONValue(document []byte) (Value, error) {
	scan, scanErr := jsonscan.Scan(document, jsonscan.MaximumNesting)
	if scanErr != nil || len(scan.DuplicatePointers()) != 0 {
		return Value{}, fmt.Errorf("resolved REST Query JSON value is invalid")
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil || bytes.Equal(canonical, []byte("null")) || !validJSONStrings(canonical) {
		return Value{}, fmt.Errorf("resolved REST Query JSON value is invalid")
	}
	value := Value{
		kind: api.LogicalTypeJSON, state: valueStateJSON,
		json: bytes.Clone(canonical), validated: true,
	}
	if err := value.Validate(); err != nil {
		return Value{}, err
	}
	return value, nil
}

func (value Value) Kind() (api.LogicalType, error) {
	if err := value.Validate(); err != nil {
		return 0, err
	}
	return value.kind, nil
}

func (value Value) Scalar() (string, bool, error) {
	if err := value.Validate(); err != nil {
		return "", false, err
	}
	return value.scalar, value.state == valueStateScalar, nil
}

func (value Value) Boolean() (bool, bool, error) {
	if err := value.Validate(); err != nil {
		return false, false, err
	}
	return value.boolean, value.state == valueStateBoolean, nil
}

func (value Value) JSON() ([]byte, bool, error) {
	if err := value.Validate(); err != nil {
		return nil, false, err
	}
	return bytes.Clone(value.json), value.state == valueStateJSON, nil
}

func (Value) String() string                       { return valueRedaction }
func (Value) GoString() string                     { return valueRedaction }
func (value Value) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Value) MarshalJSON() ([]byte, error)         { return rejectDirectJSONMarshal() }

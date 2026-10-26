package cursor

import (
	"bytes"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func decodeWireLastValues(
	sorts []decodedSort,
	values []wireLastValue,
) ([]LastValue, bool) {
	if len(values) != len(sorts) {
		return nil, false
	}
	result := make([]LastValue, len(values))
	for index, wire := range values {
		value, ok := decodeWireLastValue(sorts[index].logicalType, wire)
		if !ok {
			return nil, false
		}
		if !value.null {
			canonical, err := canonicalValueBytes(value)
			if err != nil || len(canonical) > maximumCanonicalValueBytes {
				return nil, false
			}
		}
		result[index] = value
	}
	return result, true
}

func decodeWireLastValue(kind api.LogicalType, wire wireLastValue) (LastValue, bool) {
	if wire.Null {
		if !bytes.Equal(wire.Value, []byte("null")) {
			return LastValue{}, false
		}
		return NewNullLastValue(), true
	}
	if bytes.Equal(wire.Value, []byte("null")) {
		return LastValue{}, false
	}
	if kind == api.LogicalTypeBoolean {
		var boolean bool
		if json.Unmarshal(wire.Value, &boolean) != nil {
			return LastValue{}, false
		}
		value, err := resolved.NewBooleanValue(boolean)
		if err != nil {
			return LastValue{}, false
		}
		last, err := NewLastValue(value)
		return last, err == nil
	}
	var scalar string
	if json.Unmarshal(wire.Value, &scalar) != nil {
		return LastValue{}, false
	}
	value, err := resolved.NewScalarValue(kind, scalar)
	if err != nil {
		return LastValue{}, false
	}
	last, err := NewLastValue(value)
	return last, err == nil
}

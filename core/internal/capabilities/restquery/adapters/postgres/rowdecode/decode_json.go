package rowdecode

import (
	"bytes"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func decodeJSON(fieldType api.FieldType, driverValue any) (Value, error) {
	var document []byte
	switch value := driverValue.(type) {
	case []byte:
		document = value
	case json.RawMessage:
		document = value
	default:
		return Value{}, ErrInvalidValue
	}
	resolvedValue, err := resolved.NewJSONValue(document)
	if err != nil {
		canonical, canonicalErr := attestation.Canonicalize(document)
		if canonicalErr != nil || !bytes.Equal(canonical, []byte("null")) {
			return Value{}, ErrInvalidValue
		}
		return newValue(
			fieldType, api.LogicalTypeJSON, canonical, zeroLastValue(), false, false,
		)
	}
	canonical, present, err := resolvedValue.JSON()
	if err != nil || !present {
		return Value{}, ErrInvalidValue
	}
	return newValue(
		fieldType, api.LogicalTypeJSON, canonical, zeroLastValue(), false, false,
	)
}

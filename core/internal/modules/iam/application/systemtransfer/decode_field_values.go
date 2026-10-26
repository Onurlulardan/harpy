package systemtransfer

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func decodeFieldValues(document []byte) (publicresource.FieldValues, error) {
	var wire map[string]json.RawMessage
	if json.Unmarshal(document, &wire) != nil || len(wire) < 1 {
		return publicresource.FieldValues{}, ErrInvalidMapping
	}
	values := make(map[publicresource.ID]json.RawMessage, len(wire))
	for key, value := range wire {
		fieldID, err := publicresource.ParseID(key)
		if err != nil || len(value) == 0 {
			return publicresource.FieldValues{}, ErrInvalidMapping
		}
		values[fieldID] = append(json.RawMessage(nil), value...)
	}
	result, err := publicresource.NewFieldValues(values)
	if err != nil {
		return publicresource.FieldValues{}, ErrInvalidMapping
	}
	return result, nil
}

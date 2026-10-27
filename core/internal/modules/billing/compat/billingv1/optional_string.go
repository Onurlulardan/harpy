package billingv1

import (
	"bytes"
	"encoding/json"
)

type optionalString struct {
	Present bool
	Value   *string
}

func (value *optionalString) UnmarshalJSON(document []byte) error {
	value.Present = true
	if bytes.Equal(bytes.TrimSpace(document), []byte("null")) {
		value.Value = nil
		return nil
	}
	var decoded string
	if err := json.Unmarshal(document, &decoded); err != nil {
		return err
	}
	value.Value = &decoded
	return nil
}

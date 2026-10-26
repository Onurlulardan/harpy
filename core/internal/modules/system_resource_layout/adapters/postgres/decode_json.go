package postgres

import (
	"bytes"
	"encoding/json"
	"io"
)

func decodeStrictJSON(document []byte, target any) error {
	if len(document) == 0 || target == nil {
		return layoutUnavailable()
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return layoutUnavailable()
	}
	if _, err := decoder.Token(); err != io.EOF {
		return layoutUnavailable()
	}
	return nil
}

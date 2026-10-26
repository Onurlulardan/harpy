package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func rejectDuplicateJSONMembers(document []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errors.New("JSON has trailing data")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			nameToken, nameErr := decoder.Token()
			name, stringName := nameToken.(string)
			if nameErr != nil || !stringName {
				return errors.New("JSON object member is invalid")
			}
			if _, duplicate := seen[name]; duplicate {
				return errors.New("JSON member is duplicated")
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return errors.New("JSON delimiter is invalid")
	}
	_, err = decoder.Token()
	return err
}

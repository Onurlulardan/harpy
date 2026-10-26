package schemaexecutor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func rejectDuplicateMembers(document []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	if err := walkJSONValue(decoder); err != nil {
		return fmt.Errorf("plan JSON is not strict: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		return walkJSONObject(decoder)
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter %q", delimiter)
	}
	_, err = decoder.Token()
	return err
}

func walkJSONObject(decoder *json.Decoder) error {
	seen := map[string]struct{}{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok {
			return fmt.Errorf("object member name is invalid")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate object member %q", name)
		}
		seen[name] = struct{}{}
		if err := walkJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}

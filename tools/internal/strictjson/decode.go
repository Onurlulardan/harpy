package strictjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

func Decode(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	if err := validateExactMembers(data, reflect.TypeOf(target)); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func validateExactMembers(data []byte, valueType reflect.Type) error {
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	switch valueType.Kind() {
	case reflect.Struct:
		fields := exactFields(valueType)
		members := map[string]json.RawMessage{}
		if err := json.Unmarshal(data, &members); err != nil {
			return err
		}
		for name, raw := range members {
			fieldType, exists := fields[name]
			if !exists {
				return fmt.Errorf("member %q is not an exact schema key", name)
			}
			if err := validateExactMembers(raw, fieldType); err != nil {
				return fmt.Errorf("member %q: %w", name, err)
			}
		}
	case reflect.Slice, reflect.Array:
		if valueType.Elem().Kind() == reflect.Uint8 {
			return nil
		}
		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		for index, raw := range items {
			if err := validateExactMembers(raw, valueType.Elem()); err != nil {
				return fmt.Errorf("index %d: %w", index, err)
			}
		}
	case reflect.Map:
		var members map[string]json.RawMessage
		if err := json.Unmarshal(data, &members); err != nil {
			return err
		}
		for name, raw := range members {
			if err := validateExactMembers(raw, valueType.Elem()); err != nil {
				return fmt.Errorf("map member %q: %w", name, err)
			}
		}
	}
	return nil
}

func exactFields(valueType reflect.Type) map[string]reflect.Type {
	fields := map[string]reflect.Type{}
	for index := 0; index < valueType.NumField(); index++ {
		field := valueType.Field(index)
		if field.PkgPath != "" {
			continue
		}
		tag := strings.Split(field.Tag.Get("json"), ",")[0]
		if tag == "-" {
			continue
		}
		if field.Anonymous && tag == "" {
			nested := field.Type
			for nested.Kind() == reflect.Pointer {
				nested = nested.Elem()
			}
			if nested.Kind() == reflect.Struct {
				for name, nestedType := range exactFields(nested) {
					fields[name] = nestedType
				}
				continue
			}
		}
		if tag == "" {
			tag = field.Name
		}
		fields[tag] = field.Type
	}
	return fields
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func scanValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate key %q", key)
			}
			seen[key] = true
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter %q", delimiter)
	}
	_, err = decoder.Token()
	return err
}

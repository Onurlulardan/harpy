package attestation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

func decodeStrict(data []byte, target any) error {
	return decodeStrictBounded(data, target, maximumDocumentBytes)
}

func decodeStrictBounded(data []byte, target any, maximumBytes int) error {
	if !withinDocumentSizeLimit(len(data), maximumBytes) || !utf8.Valid(data) {
		return fmt.Errorf("attestation document size is invalid")
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return fmt.Errorf("attestation JSON is not strict: %w", err)
	}
	targetType := reflect.TypeOf(target)
	if targetType == nil || targetType.Kind() != reflect.Pointer || reflect.ValueOf(target).IsNil() {
		return fmt.Errorf("attestation decode target must be a non-nil pointer")
	}
	if err := validateExactMember(data, targetType.Elem(), false); err != nil {
		return fmt.Errorf("attestation JSON shape is invalid: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode attestation JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("attestation JSON has trailing value")
	}
	return nil
}

func validateExactMember(data []byte, valueType reflect.Type, allowNull bool) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		if allowNull && (valueType.Kind() == reflect.Pointer || valueType == rawMessageType) {
			return nil
		}
		return errJSONNull
	}
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	if valueType == timeType {
		var value string
		return json.Unmarshal(data, &value)
	}
	switch valueType.Kind() {
	case reflect.Struct:
		fields := exactFields(valueType)
		members := map[string]json.RawMessage{}
		if err := json.Unmarshal(data, &members); err != nil {
			return err
		}
		for name, raw := range members {
			field, exists := fields[name]
			if !exists {
				return fmt.Errorf("member %q is not an exact contract key", name)
			}
			if err := validateExactMember(raw, field.valueType, field.allowNull); err != nil {
				return fmt.Errorf("member %q: %w", name, err)
			}
		}
		for name := range fields {
			if _, exists := members[name]; !exists && !fields[name].optional {
				return fmt.Errorf("required member %q is missing", name)
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
			if err := validateExactMember(raw, valueType.Elem(), false); err != nil {
				return fmt.Errorf("index %d: %w", index, err)
			}
		}
	}
	return nil
}

var rawMessageType = reflect.TypeOf(json.RawMessage(nil))
var timeType = reflect.TypeOf(time.Time{})

type exactField struct {
	valueType reflect.Type
	allowNull bool
	optional  bool
}

func exactFields(valueType reflect.Type) map[string]exactField {
	fields := make(map[string]exactField, valueType.NumField())
	for index := 0; index < valueType.NumField(); index++ {
		field := valueType.Field(index)
		if field.PkgPath != "" {
			continue
		}
		if field.Anonymous && field.Tag.Get("json") == "" {
			embedded := field.Type
			for embedded.Kind() == reflect.Pointer {
				embedded = embedded.Elem()
			}
			if embedded.Kind() == reflect.Struct {
				for name, exact := range exactFields(embedded) {
					fields[name] = exact
				}
				continue
			}
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		fields[name] = exactField{
			valueType: field.Type,
			allowNull: field.Tag.Get("harpyNull") == "allowed",
			optional:  field.Tag.Get("harpyOptional") == "allowed",
		}
	}
	return fields
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
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
	if token == nil {
		return nil
	}
	if number, ok := token.(json.Number); ok {
		return validateJSONNumber(number.String())
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delimiter == '{' {
			keyToken, keyErr := decoder.Token()
			if keyErr != nil {
				return keyErr
			}
			key, ok := keyToken.(string)
			if !ok || seen[key] {
				return fmt.Errorf("duplicate or invalid object key")
			}
			seen[key] = true
		}
		if err := scanValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

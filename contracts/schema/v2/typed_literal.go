package schemav2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const (
	maximumTypedLiteralBytes = 1 << 20
	maximumLiteralElements   = 10_000
	timestampLiteralLayout   = "2006-01-02T15:04:05.000000Z"
)

var (
	integerLiteralPattern = regexp.MustCompile(`^(0|-[1-9][0-9]*|[1-9][0-9]*)$`)
	decimalLiteralPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)
)

func ValidateTypedLiteralV2(literal *TypedLiteral) error {
	if literal == nil {
		return nil
	}
	if len(literal.Value) < 1 || len(literal.Value) > maximumTypedLiteralBytes ||
		bytes.Equal(literal.Value, []byte("null")) {
		return fmt.Errorf("typed literal value is empty, null, or too large")
	}
	switch literal.StorageType {
	case StorageText, StorageEnum:
		return validateLiteralText(literal.Value)
	case StorageInteger:
		return validateLiteralInteger(literal.Value)
	case StorageDecimal:
		return validateLiteralDecimal(literal.Value)
	case StorageBoolean:
		if string(literal.Value) != "true" && string(literal.Value) != "false" {
			return fmt.Errorf("boolean literal is not canonical")
		}
		return nil
	case StorageDate:
		return validateLiteralTime(literal.Value, time.DateOnly)
	case StorageTimestamp:
		return validateLiteralTime(literal.Value, timestampLiteralLayout)
	case StorageUUID:
		return validateLiteralUUID(literal.Value)
	case StorageTextList:
		return validateLiteralList(literal.Value, false)
	case StorageUUIDList:
		return validateLiteralList(literal.Value, true)
	case StorageJSON:
		canonical, err := canonicaljsonv1.Canonicalize(literal.Value, maximumTypedLiteralBytes)
		if err != nil || !bytes.Equal(canonical, literal.Value) {
			return fmt.Errorf("JSON literal is not canonical")
		}
		return nil
	default:
		return fmt.Errorf("typed literal storage type is unsupported")
	}
}

func validateLiteralText(raw []byte) error {
	value, err := literalString(raw)
	if err != nil || strings.ContainsRune(value, '\x00') || norm.NFC.String(value) != value {
		return fmt.Errorf("text literal is not canonical")
	}
	return nil
}

func validateLiteralInteger(raw []byte) error {
	value, err := literalString(raw)
	if err != nil || !integerLiteralPattern.MatchString(value) {
		return fmt.Errorf("integer literal is not canonical")
	}
	if _, err = strconv.ParseInt(value, 10, 64); err != nil {
		return fmt.Errorf("integer literal is outside bigint")
	}
	return nil
}

func validateLiteralDecimal(raw []byte) error {
	value, err := literalString(raw)
	if err != nil || len(value) > 1002 || !decimalLiteralPattern.MatchString(value) ||
		strings.HasPrefix(value, "-") && strings.Trim(value[1:], "0.") == "" {
		return fmt.Errorf("decimal literal is not canonical")
	}
	return nil
}

func validateLiteralTime(raw []byte, layout string) error {
	value, err := literalString(raw)
	if err != nil {
		return fmt.Errorf("temporal literal is not a canonical string")
	}
	parsed, err := time.Parse(layout, value)
	if err != nil || parsed.Format(layout) != value || parsed.Year() < 1 || parsed.Year() > 9999 {
		return fmt.Errorf("temporal literal is not canonical")
	}
	return nil
}

func validateLiteralUUID(raw []byte) error {
	value, err := literalString(raw)
	parsed, parseErr := uuid.Parse(value)
	if err != nil || parseErr != nil || parsed == uuid.Nil || parsed.String() != value {
		return fmt.Errorf("UUID literal is not canonical")
	}
	return nil
}

func validateLiteralList(raw []byte, uuids bool) error {
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil || len(values) > maximumLiteralElements {
		return fmt.Errorf("list literal is invalid")
	}
	for _, value := range values {
		var err error
		if uuids {
			err = validateLiteralUUID(value)
		} else {
			err = validateLiteralText(value)
		}
		if err != nil {
			return err
		}
	}
	canonical, _ := json.Marshal(values)
	if !bytes.Equal(canonical, raw) {
		return fmt.Errorf("list literal is not canonical")
	}
	return nil
}

func literalString(raw []byte) (string, error) {
	var value string
	if json.Unmarshal(raw, &value) != nil || !canonicalString(raw, value) {
		return "", fmt.Errorf("literal is not a canonical string")
	}
	return value, nil
}

func canonicalString(raw []byte, value string) bool {
	encoded, err := json.Marshal(value)
	return err == nil && bytes.Equal(encoded, raw)
}

package effectsv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

var (
	lowerSHA256   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	sqlName       = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	effectKind    = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
	providerName  = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
	operationName = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
	machineCode   = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)
)

func decodeStrict(document []byte, target any) error {
	if len(bytes.TrimSpace(document)) == 0 || len(document) > MaximumDocumentBytes || !utf8.Valid(document) {
		return fmt.Errorf("document size or UTF-8 is invalid")
	}
	if err := canonicaljsonv1.Validate(document); err != nil {
		return fmt.Errorf("document JSON is ambiguous: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("document has trailing content")
	}
	return nil
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validText(value string) bool {
	return utf8.ValidString(value) && value != "" && strings.TrimSpace(value) == value
}

func validBoundedText(value string, maximumBytes int) bool {
	return validText(value) && len([]byte(value)) <= maximumBytes
}

func validCorrelationID(value string) bool {
	if len(value) < 1 || len(value) > MaximumCorrelationIDBytes || !asciiAlphaNumeric(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if !asciiAlphaNumeric(character) && character != '.' && character != '_' &&
			character != ':' && character != '-' {
			return false
		}
	}
	return true
}

func asciiAlphaNumeric(character byte) bool {
	return character >= 'A' && character <= 'Z' ||
		character >= 'a' && character <= 'z' ||
		character >= '0' && character <= '9'
}

func validTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed, err == nil && parsed.Location() == time.UTC &&
		parsed.Format(time.RFC3339Nano) == value
}

func cloneOptional(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func equalOptional(left *string, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
